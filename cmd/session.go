/*
Copyright © 2024 github
*/
package cmd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type owner_repo struct {
	owner string
	repo  string
}

type b64gztar struct {
	tgz_filepath string
}

type db_location interface {
	DBPATH() string
}

type db_location_local struct {
	db_location
	prefix string
	db_dir string
}

func (dbl db_location_local) DBPATH() string {
	return filepath.Join(dbl.prefix, dbl.db_dir)
}

type MirvaSession struct {
	id              int
	owner           string
	controller_repo string

	query_pack   b64gztar
	language     string
	repositories []owner_repo

	skipped_repos  []skipped_repo_element
	analysis_repos map[owner_repo]db_location
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

func (sn MirvaSession) submit_response(w http.ResponseWriter) {
	// TODO
}

func (sn MirvaSession) start_analyses() {
	// TODO
}

func (sn MirvaSession) collect_info(w http.ResponseWriter, r *http.Request) {
	// TODO: collect:
	// 2024/02/14 10:20:13     "language": "cpp"
	// 2024/02/14 10:20:13     "repositories": "[google/flatbuffers]"
	// 2024/02/14 10:20:13     "query_pack":
	// 2024/02/14 10:20:13         base64 encoded gzipped tar file, contents {
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
			slog.Error("Invalid owner / repo entry", "entry", t)
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		sn.repositories = append(sn.repositories, owner_repo{t[0], t[1]})
	}

	sn.save()

}

func (sn MirvaSession) extract_tgz(qp string) error {
	// These are decoded manually via
	//    base64 -d < foo1 | gunzip | tar t | head -20
	// base64 decode the body
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
	fpath := path.Join(cwd, fmt.Sprintf("querypack-%d.tgz", sn.id))
	err = os.WriteFile(fpath, tgz, 0644)
	if err != nil {
		slog.Error("unable to save querypack body decoding error", "path", fpath)
		return err
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

// For every repo with a built database we ultimately
// run the queries to create the sarif file:
//
// cd ~/local
// codeql database analyze --format=sarif-latest --rerun \
// 	   --output $QUERYNAME.sarif \
// 	   --search-path $QLLIB \
// 	   -j8 \
// 	   -- $DBPATH $QUERYPACKS

func (sn MirvaSession) save() {
	// sqlite state retention
	// TODO
}

func (sn MirvaSession) load() {
	// sqlite state retention
	// TODO
}

func (sn MirvaSession) find_available_DBs() {
	sn.load()

	// TODO: Determine for which repositories codeql databases are available.
	// Those will be the analysis_repos.  The rest will be skipped.
	//
	// skipped_repos  []skipped_repo_element
	// analysis_repos map[owner_repo]db_location

	// sn.analysis_repos
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

// This interface flattens the tree; it may be easier to use a
// []skipped_repo_element in go, then reproduce the tree in json.
type skipped_repo_element interface {
	Reason() string
	Count_Key() string
	Count() int
	Repositories_Key() string
	Resitories() []owner_repo
}

type unused_repo struct {
	count_key        string
	count            int
	repositories_key string
	repositories     []string
}
type access_mismatch_repos struct {
	unused_repo
}
type not_found_repos struct {
	unused_repo
}
type no_codeql_db_repos struct {
	unused_repo
}
type over_limit_repos struct {
	unused_repo
}

func (u unused_repo) Count_Key() string {
	return u.count_key
}
func (u unused_repo) Count() int {
	return u.count
}
func (u unused_repo) Repositories_Key() string {
	return u.repositories_key
}
func (u unused_repo) Repositories() []string {
	return u.repositories
}

func (_ access_mismatch_repos) Reason() string {
	return "access_mismatch_repos"
}
func (_ not_found_repos) Reason() string {
	return "not_found_repos"
}
func (_ no_codeql_db_repos) Reason() string {
	return "no_codeql_db_repos"
}
func (_ over_limit_repos) Reason() string {
	return "over_limit_repos"
}
