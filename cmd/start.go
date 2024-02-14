/*
Copyright © 2024 github
*/
package cmd

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
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

	// Trigger a new MRVA run
	// POST https://api.github.com/repos/hohn/mirva-controller/code-scanning/codeql/variant-analyses
	r.HandleFunc("/repos/{owner}/{repo}/code-scanning/codeql/variant-analyses", NewMirvaOwRe)
	// 			  /repos/hohn   /mirva-controller/code-scanning/codeql/variant-analyses
	// Or via
	r.HandleFunc("/{repository_id}/code-scanning/codeql/variant-analyses", NewMirvaId)

	r.HandleFunc("/", RootHandler)

	// Bind to a port and pass our router in
	log.Fatal(http.ListenAndServe(":8080", r))
}

func RootHandler(w http.ResponseWriter, r *http.Request) {
	LogAbove(LogWarning, "Request on /")
}

func NewMirvaId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	LogAbove(LogWarning, "New mrva using repository_id=%v\n", vars["repository_id"])
}

func NewMirvaOwRe(w http.ResponseWriter, r *http.Request) {
	LogAbove(LogWarning, "New mrva run from owner/repo\n")
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
