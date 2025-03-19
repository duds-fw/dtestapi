package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dtestapi",
	Short: "CLI tool for API testing",
	Long:  "dtestapi allows developers to run API tests from JSON files, with support for dependencies, parallel execution, and logging.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
