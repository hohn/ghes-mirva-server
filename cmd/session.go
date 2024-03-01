/*
Copyright © 2024 github
*/
package cmd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/hohn/ghes-mirva-server/analyze"
)

type owner_repo_loc struct {
	owner    string
	repo     string
	location db_location
}

type b64gztar struct {
	tgz_filepath string
}

type db_location interface {
	DBPATH() string
}

type db_location_local struct {
	prefix  string
	db_file string
}

func (dbl db_location_local) DBPATH() string {
	return filepath.Join(dbl.prefix, dbl.db_file)
}

type MirvaSession struct {
	id              int
	owner           string
	controller_repo string

	query_pack   b64gztar
	language     string
	repositories []owner_repo_loc

	access_mismatch_repos access_mismatch_repos
	not_found_repos       not_found_repos
	no_codeql_db_repos    no_codeql_db_repos
	over_limit_repos      over_limit_repos

	analysis_repos map[owner_repo_loc]db_location
}

func _next_id() func() int {
	// Get a 5+ digit id
	current_id := 93521
	return func() int {
		current_id += 1
		return current_id
	}
}

var next_id = _next_id()

func (sn *MirvaSession) submit_response(w http.ResponseWriter) {

	slog.Debug("Forming and sending response for submitted analysis job", "id", sn.id)

	// Construct the response bottom-up
	// TODO functional style
	var m_cr ControllerRepo
	var m_ac Actor

	r_nfr := sn.arr_to_json_NFR()

	r_amr := sn.arr_to_json_AMR()

	r_ncd := sn.arr_to_json_NCDB()

	// TODO fill these
	r_olr := sn.arr_to_json_OLR()

	m_skip := SkippedRepositories{r_amr, r_nfr, r_ncd, r_olr}

	var m_sr SubmitResponse
	m_sr.Actor = m_ac
	m_sr.ControllerRepo = m_cr
	m_sr.ID = sn.id
	m_sr.QueryLanguage = sn.language
	m_sr.QueryPackURL = sn.query_pack.tgz_filepath
	m_sr.CreatedAt = time.Now().Format(time.RFC3339)
	m_sr.UpdatedAt = time.Now().Format(time.RFC3339)
	m_sr.Status = "in_progress"
	m_sr.SkippedRepositories = m_skip

	// Encode the response as JSON
	submit_response, err := json.Marshal(m_sr)
	if err != nil {
		log.Println("Error encoding response as JSON:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send analysisReposJSON via ResponseWriter
	w.Header().Set("Content-Type", "application/json")
	w.Write(submit_response)

}

/*
See macros.m4 for generating these
*/
func (sn *MirvaSession) arr_to_json_NCDB() NoCodeqlDBRepos {
	var r__ NoCodeqlDBRepos
	r__.Repositories = []string{}
	r__.RepositoryCount = len(sn.no_codeql_db_repos.orl)
	for _, repo := range sn.no_codeql_db_repos.orl {
		r__.Repositories = append(r__.Repositories,
			fmt.Sprintf("%s/%s", repo.owner, repo.repo))
	}
	return r__
}

func (sn *MirvaSession) arr_to_json_AMR() AccessMismatchRepos {
	var r__ AccessMismatchRepos
	r__.Repositories = []string{}
	r__.RepositoryCount = len(sn.access_mismatch_repos.orl)
	for _, repo := range sn.access_mismatch_repos.orl {
		r__.Repositories = append(r__.Repositories,
			fmt.Sprintf("%s/%s", repo.owner, repo.repo))
	}
	return r__
}

func (sn *MirvaSession) arr_to_json_OLR() OverLimitRepos {
	var r__ OverLimitRepos
	r__.Repositories = []string{}
	r__.RepositoryCount = len(sn.over_limit_repos.orl)
	for _, repo := range sn.over_limit_repos.orl {
		r__.Repositories = append(r__.Repositories,
			fmt.Sprintf("%s/%s", repo.owner, repo.repo))
	}
	return r__
}

func (sn *MirvaSession) arr_to_json_NFR() NotFoundRepos {
	var r__ NotFoundRepos
	r__.RepositoryFullNames = []string{}
	r__.RepositoryCount = len(sn.not_found_repos.orl)
	for _, repo := range sn.not_found_repos.orl {
		r__.RepositoryFullNames = append(r__.RepositoryFullNames,
			fmt.Sprintf("%s/%s", repo.owner, repo.repo))
	}
	return r__
}

/*
end generated code
*/

func (sn *MirvaSession) start_analyses() {
	slog.Debug("Queueing codeql database analyze jobs")

	for orl := range sn.analysis_repos {
		info := analyze.AnalyzeJob{
			QueryPackId:   sn.id,
			QueryLanguage: sn.language,

			DBOwner: orl.owner,
			DBRepo:  orl.repo,
		}
		analyze.Jobs <- info
	}
}

// Collect the following info from the request body
//
//	"language": "cpp"
//	"repositories": "[google/flatbuffers]"
//	"query_pack":
//	    base64 encoded gzipped tar file, contents {...}
func (sn *MirvaSession) collect_info(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Collecting session info")

	if r.Body == nil {
		err := "Missing request body"
		log.Println(err)
		http.Error(w, err, http.StatusNoContent)
		return
	}
	buf, err := io.ReadAll(r.Body)
	if err != nil {
		var w http.ResponseWriter
		slog.Error("Error reading MRVA submission body", "error", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	msg, err := TrySubmitMsg(buf)
	if err != nil {
		// Unknown message
		slog.Error("Unknown MRVA submission body format")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Decompose the SubmitMsg and keep information in the MirvaSession

	// 1. Save the query pack and keep the location
	if !is_base64_gzip([]byte(msg.QueryPack)) {
		slog.Error("MRVA submission body querypack has invalid format")
		err := errors.New("MRVA submission body querypack has invalid format")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = sn.extract_tgz(msg.QueryPack)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 2. Save the language
	sn.language = msg.Language

	// 3. Save the repositories
	for _, v := range msg.Repositories {
		t := strings.Split(v, "/")
		if len(t) != 2 {
			slog.Error("Invalid owner / repository entry", "entry", t)
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		sn.repositories = append(sn.repositories, owner_repo_loc{t[0], t[1], nil})
	}

	sn.save()

}

func (sn *MirvaSession) extract_tgz(qp string) error {
	// These are decoded manually via
	//    base64 -d < foo1 | gunzip | tar t | head -20
	// base64 decode the body
	slog.Debug("Extracting query pack")

	tgz, err := base64.StdEncoding.DecodeString(qp)
	if err != nil {
		slog.Error("querypack body decoding error:", err)
		return err
	}
	// Save the tar.gz body
	cwd, err := os.Getwd()
	if err != nil {
		slog.Error("No working directory")
		panic(err)
	}

	dirpath := path.Join(cwd, "var", "codeql", "querypacks")
	if err := os.MkdirAll(dirpath, 0755); err != nil {
		slog.Error("Unable to create query pack output directory",
			"dir", dirpath)
		return err
	}

	fpath := path.Join(dirpath, fmt.Sprintf("qp-%d.tgz", sn.id))
	err = os.WriteFile(fpath, tgz, 0644)
	if err != nil {
		slog.Error("unable to save querypack body decoding error", "path", fpath)
		return err
	} else {
		slog.Info("Query pack saved to ", "path", fpath)
	}
	sn.query_pack.tgz_filepath = fpath
	return nil
}

func is_base64_gzip(val []byte) bool {
	// Some important payloads can be listed via
	// base64 -d < foo1 | gunzip | tar t|head -20
	//
	// This function checks the request body up to the `gunzip` part.
	//
	if len(val) >= 4 {
		// Extract header
		hdr := make([]byte, base64.StdEncoding.DecodedLen(4))
		_, err := base64.StdEncoding.Decode(hdr, []byte(val[0:4]))
		if err != nil {
			log.Println("WARNING: IsBase64Gzip decode error:", err)
			return false
		}
		// Check for gzip heading
		magic := []byte{0x1f, 0x8b}
		if bytes.Equal(hdr[0:2], magic) {
			return true
		} else {
			return false
		}
	} else {
		return false
	}
}

type SubmitMsg struct {
	ActionRepoRef string   `json:"action_repo_ref"`
	Language      string   `json:"language"`
	QueryPack     string   `json:"query_pack"`
	Repositories  []string `json:"repositories"`
}

// See if the buf contains a json-encoded message
func TrySubmitMsg(buf []byte) (SubmitMsg, error) {
	buf1 := make([]byte, len(buf))
	copy(buf1, buf)
	dec := json.NewDecoder(bytes.NewReader(buf1))
	dec.DisallowUnknownFields()
	var m SubmitMsg
	err := dec.Decode(&m)
	return m, err
}

func (sn *MirvaSession) save() {
	// TODO sqlite state retention

}

func (sn *MirvaSession) load() {
	// TODO sqlite state retention
}

//		Determine for which repositories codeql databases are available.
//
//	 Those will be the analysis_repos.  The rest will be skipped.
func (sn *MirvaSession) find_available_DBs() {

	slog.Debug("Looking for available CodeQL databases")

	sn.load()

	cwd, err := os.Getwd()
	if err != nil {
		slog.Error("No working directory")
		return
	}

	if sn.analysis_repos == nil {
		sn.analysis_repos = map[owner_repo_loc]db_location{}
	}

	// We're looking for paths like
	// codeql/sarif/google/flatbuffers/google_flatbuffers.sarif
	for _, rep := range sn.repositories {
		dbPrefix := filepath.Join(cwd, "codeql", "dbs", rep.owner, rep.repo)
		dbName := fmt.Sprintf("%s_%s_db.zip", rep.owner, rep.repo)
		dbPath := filepath.Join(dbPrefix, dbName)

		if _, err := os.Stat(dbPath); errors.Is(err, fs.ErrNotExist) {
			slog.Info("Database does not exist for repository ", "owner/repo", rep,
				"path", dbPath)
			sn.not_found_repos.orl = append(sn.not_found_repos.orl, rep)
		} else {
			slog.Info("Found database for ", "owner/repo", rep, "path", dbPath)
			sn.analysis_repos[rep] = db_location_local{dbPrefix, dbName}
		}
	}

	sn.save()
}

// Define types to represent the json map
// 	"skipped_repositories": {
// 		"access_mismatch_repos": {
// 			"repository_count": 0,
// 			"repositories": [] },
// 		"not_found_repos": {
// 			"repository_count": 0,
// 			"repository_full_names": []
// 		},
// 		"no_codeql_db_repos": {
// 			"repository_count": 0,
// 			"repositories": []
// 		},
// 		"over_limit_repos": {
// 			"repository_count": 0,
// 			"repositories": []
// 		}
// 	}

// This interface flattens the tree; it is easier to use 4 typed arrays
// in Go and then form the tree before producing json.
type skipped_repo_element interface {
	Reason() string
	Count_Key() string
	Count() int
	Repository_Key() string
	Repository() owner_repo_loc
}

type access_mismatch_repos struct {
	orl []owner_repo_loc
}

type no_codeql_db_repos struct {
	orl []owner_repo_loc
}

type over_limit_repos struct {
	orl []owner_repo_loc
}

type not_found_repos struct {
	orl []owner_repo_loc
}

func (n not_found_repos) Reason() string {
	return "not_found_repos"
}

func (n not_found_repos) Count_Key() string {
	return "not_found_repos"
}

func (n not_found_repos) Count() int {
	return len(n.orl)
}

func (n not_found_repos) Repository_Key() string {
	return "repository_full_names"
}

func (n not_found_repos) Repository() owner_repo_loc {
	return n.orl[0]
}

func (u over_limit_repos) Count() int {
	return len(u.orl)
}
func (u over_limit_repos) Repository_Key() string {
	return "over_limit_repos"
}
func (u over_limit_repos) Repository() owner_repo_loc {
	return u.orl[0]
}

func (_ access_mismatch_repos) Reason() string {
	return "access_mismatch_repos"
}

func (_ no_codeql_db_repos) Reason() string {
	return "no_codeql_db_repos"
}
