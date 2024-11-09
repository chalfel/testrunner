package cmd

import (
	"fmt"
	"os"

	"github.com/chalfel/testrunner/runner"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "test-runner",
	Short: "Parallel test runner with PostgreSQL containers",
	Run: func(cmd *cobra.Command, args []string) {
		config := runner.Config{
			TestFolder:  testFolder,
			BlockSize:   blockSize,
			TestCommand: testCommand,
			BasePort:    basePort,
			FilePattern: filePattern,
			Verbose:     verbose,
		}
		runner.RunTestBatches(config)
	},
}

var (
	testFolder  string
	blockSize   int
	testCommand string
	basePort    int
	filePattern string
	verbose     bool
)

func init() {
	rootCmd.Flags().StringVarP(&testFolder, "test-folder", "f", ".", "Path to the folder containing test files")
	rootCmd.Flags().IntVarP(&blockSize, "block-size", "b", 25, "Number of test files per PostgreSQL container")
	rootCmd.Flags().StringVarP(&testCommand, "test-command", "c", "go test", "Command to execute each test file")
	rootCmd.Flags().IntVarP(&basePort, "base-port", "p", 5433, "Base port for PostgreSQL containers")
	rootCmd.Flags().StringVarP(&filePattern, "file-pattern", "P", "*_test.go", "File pattern to match test files")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
