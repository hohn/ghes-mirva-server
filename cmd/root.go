/*
Copyright © 2024 github

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"log"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ghes-mirva-server",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) {
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		switch logging_verbosity {
		case "debug":
			slog.SetLogLoggerLevel(slog.LevelDebug)
		case "info":
			slog.SetLogLoggerLevel(slog.LevelInfo)
		case "warn":
			slog.SetLogLoggerLevel(slog.LevelWarn)
		case "error":
			slog.SetLogLoggerLevel(slog.LevelError)
		default:
			log.Printf("Invalid logging verbosity level: %s", logging_verbosity)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ghes-mirva-server.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.Flags().StringVar(&Backend, "backend", "local", `Backend to use.
	Currently available:
	- local
	`)

	rootCmd.PersistentFlags().StringVar(&logging_verbosity, "verbosity", "info", `Logging verbosity, from least to most verbose: 
	- error
	- warn
	- info
	- debug
	`)
}

var logging_verbosity string

var Backend string = "local"

const (
	LogUnused int = iota
	LogFatal
	LogError
	LogWarning
	LogInfo
	LogTrace
	LogAll
)

var Verbosity int

func LogAbove(level int, format string, args ...interface{}) {
	if Verbosity > level {
		log.Printf(format, args...)
	}
}
