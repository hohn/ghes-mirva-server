dnl Some quick and dirty macros to work around missing template functionality

define(`arr_to_json', `
func (sn MirvaSession) arr_to_json1(r_ncd NoCodeqlDBRepos) {
	r_ncd.RepositoryCount = len(sn.no_codeql_db_repos.ndb)
	for _, repo := range sn.no_codeql_db_repos.ndb {
		r_ncd.Repositories = append(r_ncd.Repositories,
			fmt.Sprintf("%s/%s", repo.owner, repo.repo))
	}
}
')
arr_to_json()

define(`arr_to_json', `
func (sn MirvaSession) arr_to_json_$3(r_ncd $1) {
	r_ncd.RepositoryCount = l`'en(sn.$2.orl)
	for _, repo := range sn.$2.orl {
		r_ncd.Repositories = append(r_ncd.Repositories,
			fmt.Sprintf("%s/%s", repo.owner, repo.repo))
	}
}
')
arr_to_json(NoCodeqlDBRepos,     no_codeql_db_repos,    NCDB)
arr_to_json(AccessMismatchRepos, access_mismatch_repos, AMR)
arr_to_json(OverLimitRepos,      over_limit_repos,      OLR)
