/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/vishu42/terrasome/cmd/cli/cmd/run"
)

// versionCmd represents the plan command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "displays the version of client and server",
	Run:   run.RunVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	versionCmd.PersistentFlags().String("server-addr", "", "the server address")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// versionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}