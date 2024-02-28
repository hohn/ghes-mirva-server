dnl Some quick and dirty macros to work around missing template functionality

define(`arr_to_json', `
func (sn *MirvaSession) arr_to_json1(r_ncd NoCodeqlDBRepos) {
	r_ncd.RepositoryCount = len(sn.no_codeql_db_repos.ndb)
	for _, repo := range sn.no_codeql_db_repos.ndb {
		r_ncd.Repositories = append(r_ncd.Repositories,
			fmt.Sprintf("%s/%s", repo.owner, repo.repo))
	}
}
')
arr_to_json()

define(`arr_to_json', `
func (sn *MirvaSession) arr_to_json_$3() $1 {
	var r__ $1
        r__.$4 = []string{}
	r__.RepositoryCount = l`'en(sn.$2.orl)
	for _, repo := range sn.$2.orl {
		r__.$4 = append(r__.$4,
			fmt.Sprintf("%s/%s", repo.owner, repo.repo))
	}
        return r__
}
')
arr_to_json(NoCodeqlDBRepos,     no_codeql_db_repos,    NCDB , Repositories)
arr_to_json(AccessMismatchRepos, access_mismatch_repos, AMR  , Repositories)
arr_to_json(OverLimitRepos,      over_limit_repos,      OLR  , Repositories)
arr_to_json(NotFoundRepos,       not_found_repos,       NFR  , RepositoryFullNames)
