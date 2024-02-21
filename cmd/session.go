/*
Copyright © 2024 github
*/
package cmd

import (
	"net/http"
	"path/filepath"
)

type owner_repo struct {
	owner string
	repo  string
}

type b64gztar struct {
	raw string
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
	owner           string
	controller_repo string
	language        string
	repositories    []owner_repo
	query_pack      b64gztar
	skipped_repos   []skipped_repo_element
	analysis_repos  map[owner_repo]db_location
}

func (sn MirvaSession) collect_info(r *http.Request) {
	// TODO: collect:
	// 2024/02/14 10:20:13     "language": "cpp"
	// 2024/02/14 10:20:13     "repositories": "[google/flatbuffers]"
	// 2024/02/14 10:20:13     "query_pack":
	// 2024/02/14 10:20:13         base64 encoded gzipped tar file, contents {

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
}

func (sn MirvaSession) find_available_DBs() (map[owner_repo]db_location, error) {
	// TODO
	err := error(nil)
	return sn.analysis_repos, err
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

func (ms *MirvaSession) find_available_DBs() {
	//

}
