package runner

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// Config holds the configuration for the test runner
type Config struct {
	TestFolder  string
	BlockSize   int
	TestCommand string
	BasePort    int
	FilePattern string
	Verbose     bool
}

// RunTestBatches initializes and runs tests in parallel batches
func RunTestBatches(config Config) {
	var testFiles []string
	err := filepath.Walk(config.TestFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		matched, _ := filepath.Match(config.FilePattern, info.Name())
		if !info.IsDir() && matched {
			testFiles = append(testFiles, path)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Failed to find test files: %v", err)
	}

	// Run tests in batches
	var wg sync.WaitGroup
	for i := 0; i < len(testFiles); i += config.BlockSize {
		batch := testFiles[i:min(i+config.BlockSize, len(testFiles))]
		containerIndex := i/config.BlockSize + 1
		containerName := fmt.Sprintf("postgres_test_%d", containerIndex)
		port := config.BasePort + containerIndex // Increment port based on block

		wg.Add(1)
		go runTestBatch(batch, config.TestCommand, containerName, port, config.Verbose, &wg)
	}

	wg.Wait()
	fmt.Println("All test batches completed.")
}

func runTestBatch(batchFiles []string, testCommand, containerName string, port int, verbose bool, wg *sync.WaitGroup) {
	defer wg.Done()

	// Start PostgreSQL container for this batch with a unique port
	if err := startPostgresContainer(containerName, port, verbose); err != nil {
		log.Fatalf("Failed to start container %s on port %d: %v", containerName, port, err)
	}
	defer cleanupContainer(containerName, verbose)

	// Wait for PostgreSQL to be ready
	if verbose {
		fmt.Printf("Container %s started on port %d. Waiting for PostgreSQL to be ready...\n", containerName, port)
	}
	exec.Command("sleep", "5").Run()

	// Run tests
	for _, testFile := range batchFiles {
		cmd := exec.Command("sh", "-c", fmt.Sprintf("%s %s", testCommand, testFile))
		cmd.Env = append(os.Environ(),
			"POSTGRES_HOST=localhost",
			"POSTGRES_PORT="+strconv.Itoa(port),
			"POSTGRES_USER=test",
			"POSTGRES_PASSWORD=test",
			"POSTGRES_DB=testdb",
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if verbose {
			fmt.Printf("Running command: %s %s\n", testCommand, testFile)
		}

		if err := cmd.Run(); err != nil {
			log.Printf("Test failed for file %s: %v", testFile, err)
			return
		}
	}

	fmt.Printf("Completed batch in container %s on port %d\n", containerName, port)
}

func startPostgresContainer(containerName string, port int, verbose bool) error {
	cmd := exec.Command("docker", "run", "--name", containerName, "-e", "POSTGRES_USER=test", "-e", "POSTGRES_PASSWORD=test", "-e", "POSTGRES_DB=testdb", "-p", fmt.Sprintf("%d:5432", port), "-d", "postgres")
	if verbose {
		fmt.Printf("Starting PostgreSQL container with command: %s\n", strings.Join(cmd.Args, " "))
	}
	return cmd.Run()
}

func cleanupContainer(containerName string, verbose bool) {
	if verbose {
		fmt.Printf("Stopping and removing container: %s\n", containerName)
	}
	exec.Command("docker", "stop", containerName).Run()
	exec.Command("docker", "rm", containerName).Run()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
