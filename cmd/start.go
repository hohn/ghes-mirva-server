/*
Copyright © 2024 github
*/
package cmd

import (
	"log"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/hohn/ghes-mirva-server/analyze"
	co "github.com/hohn/ghes-mirva-server/common"
	"github.com/hohn/ghes-mirva-server/store"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		LogAbove(LogWarning, "Starting server")
		serve()
	},
}

func serve() {
	r := mux.NewRouter()

	r.HandleFunc("/repos/{owner}/{repo}/code-scanning/codeql/variant-analyses", MirvaRequest)
	// 			  /repos/hohn   /mirva-controller/code-scanning/codeql/variant-analyses
	// Or via
	r.HandleFunc("/{repository_id}/code-scanning/codeql/variant-analyses", MirvaRequestID)

	r.HandleFunc("/", RootHandler)

	// This is the standalone status request.
	// It's also the first request made when downloading; the difference is on the
	// client side's handling.
	r.HandleFunc("/repos/{owner}/{repo}/code-scanning/codeql/variant-analyses/{codeql_variant_analysis_id}", MirvaStatus)

	r.HandleFunc("/repos/{owner}/{repo}/code-scanning/codeql/variant-analyses/{codeql_variant_analysis_id}/repos/{repo_owner}/{repo_name}", MirvaDownLoad2)

	r.HandleFunc("/codeql-query-console/codeql-variant-analysis-repo-tasks/{codeql_variant_analysis_id}/{repo_id}/{owner_id}/{controller_repo_id}", MirvaDownLoad3)

	r.HandleFunc("/github-codeql-query-console-prod/codeql-variant-analysis-repo-tasks/{codeql_variant_analysis_id}/{repo_id}", MirvaDownLoad4)

	// Bind to a port and pass our router in
	log.Fatal(http.ListenAndServe(":8080", r))
}

func RootHandler(w http.ResponseWriter, r *http.Request) {
	LogAbove(LogWarning, "Request on /")
}

func MirvaStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slog.Info("mrva status request for ",
		"owner", vars["owner"],
		"repo", vars["repo"],
		"codeql_variant_analysis_id", vars["codeql_variant_analysis_id"])
	id, err := strconv.Atoi(vars["codeql_variant_analysis_id"])
	if err != nil {
		slog.Error("Variant analysis is is not integer", "id",
			vars["codeql_variant_analysis_id"])
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	js := co.JobSpec{
		ID: id,
		Orl: co.OwnerRepoLoc{
			Owner: vars["owner"],
			Repo:  vars["repo"],
			Location: co.DBLocation{
				DBPATH: "",
			},
		},
	}

	analyze.StatusResponse(w,
		js,
		store.GetJobInfo(js),
	)
}

func MirvaDownLoad2(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	LogAbove(LogWarning, "mrva download step 2 for (%s,%s,%s,%s,%s)\n",
		vars["owner"],
		vars["repo"],
		vars["codeql_variant_analysis_id"],
		vars["repo_owner"],
		vars["repo_name"])
}

func MirvaDownLoad3(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	LogAbove(LogWarning, "mrva download step 3 for (%s,%s,%s,%s)\n",
		vars["codeql_variant_analysis_id"],
		vars["repo_id"],
		vars["owner_id"],
		vars["controller_repo_id"])
}

func MirvaDownLoad4(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	LogAbove(LogWarning, "mrva download step 4 for (%s,%s)\n",
		vars["codeql_variant_analysis_id"],
		vars["repo_id"])
}

func MirvaRequestID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	LogAbove(LogWarning, "New mrva using repository_id=%v\n", vars["repository_id"])
}

func MirvaRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slog.Info("New mrva run ", "owner", vars["owner"], "repo", vars["repo"])
	// TODO Change this to functional style?
	session := new(MirvaSession)
	session.id = next_id()
	session.owner = vars["owner"]
	session.controller_repo = vars["repo"]
	session.collect_info(w, r)
	session.find_available_DBs()
	session.start_analyses()
	session.submit_response(w)
	session.save()
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
