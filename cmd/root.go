package cmd

import (
	"github.com/spf13/cobra"
	"oproller/cmd/precompile"
	"oproller/cmd/setup"
	"oproller/cmd/version"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "oproller",
	Short: "A simple CLI tool to setup and register precompiled also pre-deployed contracts.",
	Long:  `A simple CLI tool to setup and register precompiled also pre-deployed contracts.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(version.Cmd())
	rootCmd.AddCommand(setup.Cmd())
	rootCmd.AddCommand(precompile.Cmd())
}
