package cmd

import (
	"fmt"

	"github.com/duds-fw/dtestapi/internal"
	"github.com/spf13/cobra"
)

var testFile string
var parallel bool
var outputLog string

func init() {
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run API tests",
		Run: func(cmd *cobra.Command, args []string) {
			testFile, _ := cmd.Flags().GetString("test-case")
			fmt.Println("Loading test cases from:", testFile)
			err := internal.RunTests(testFile, parallel, outputLog)
			if err != nil {
				fmt.Println("Error running tests:", err)
			}
		},
	}

	runCmd.Flags().StringVarP(&testFile, "test-case", "t", "tests.json", "Path to test case JSON file")
	runCmd.Flags().BoolVarP(&parallel, "parallel", "p", false, "Run independent tests in parallel")
	runCmd.Flags().StringVarP(&outputLog, "output", "o", "log.json", "Path to JSON log file")

	rootCmd.AddCommand(runCmd)
}
