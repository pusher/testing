package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"

	"github.com/jstemmer/go-junit-report/formatter"
	"github.com/jstemmer/go-junit-report/parser"
)

func main() {
	// First argument is the executable, don't need that
	args := os.Args[1:]
	realGo := getEnv("GO_WRAPPER_REAL_GO", "/usr/local/go/bin/go")
	junitPath := getEnv("GO_WRAPPER_JUNIT_PATH", "")
	debugEnv := getEnv("GO_WRAPPER_DEBUG", "false")
	debug, err := strconv.ParseBool(debugEnv)
	if err != nil {
		log.Printf("Expected GO_WRAPPER_DEBUG to be bool, got %s: %v", debugEnv, err)
	}

	// If not requesting Junit, don't bother streaming the output
	if junitPath == "" || len(args) < 1 || args[0] != "test" {
		if debug {
			log.Printf("Running without JUnit output: go %v", args)
		}
		runWithoutJunit(realGo, args...)
		return
	}

	if debug {
		log.Printf("Running with JUnit output: go %v", args)
	}

	// JunitPath should be set
	// First argument should be test, no need to pass first argument
	runWithJunit(realGo, junitPath, args[1:]...)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func runWithoutJunit(realGo string, args ...string) {
	// Run the actual go process with the args input
	cmd := exec.Command(realGo, args...)

	// Attach Stdout, Stderr and Stdin as normal
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Execute the command
	err := cmd.Run()

	// If theres an error, exit with the correct code if possible
	if err != nil {
		log.Printf("Error executing command: %v", err)
		exitError, ok := err.(*exec.ExitError)
		if !ok {
			os.Exit(1)
		} else {
			os.Exit(exitError.ExitCode())
		}
	}

	// No error so exit 0
	os.Exit(0)
}

func runWithJunit(realGo, junitPath string, args ...string) {
	// Make sure we have verbose test output
	if contains(args, "-v") {
		args = append([]string{"test"}, args...)
	} else {
		args = append([]string{"test", "-v"}, args...)
	}

	// Run the actual go process with the args input
	cmd := exec.Command(realGo, args...)

	// Attach Stdout, Stderr and Stdin, copying Stderr and Stdout to the buffer
	junitBuffer := bytes.NewBuffer([]byte{})
	cmd.Stdin = os.Stdin
	cmd.Stdout = &writerCopier{out: os.Stdout, copy: junitBuffer}
	cmd.Stderr = &writerCopier{out: os.Stderr, copy: junitBuffer}

	// Execute the command
	err := cmd.Run()

	// If theres an error, exit with the correct code if possible
	if err != nil {
		log.Printf("Error executing command: %v", err)
		exitError, ok := err.(*exec.ExitError)
		if !ok {
			os.Exit(1)
		} else {
			os.Exit(exitError.ExitCode())
		}
	}

	// Parse the go test result into a report
	report, err := parser.Parse(junitBuffer, "")
	if err != nil {
		log.Printf("Error parsing go test output: %v", err)
		os.Exit(1)
	}

	outFile, err := os.Create(junitPath)
	if err != nil {
		log.Printf("Error opening file %s: %v", junitPath, err)
		os.Exit(1)
	}
	defer outFile.Close()

	// Write the report to the output file
	err = formatter.JUnitReportXML(report, false, "", outFile)
	if err != nil {
		fmt.Printf("Error writing XML: %s\n", err)
		os.Exit(1)
	}

	// No error so exit 0
	os.Exit(0)
}

type writerCopier struct {
	out  io.Writer
	copy io.Writer
}

func (w *writerCopier) Write(p []byte) (int, error) {
	// Write to the main output
	n, err := w.out.Write(p)
	if err != nil {
		return n, err
	}

	// Write to the copy
	n, err = w.copy.Write(p)
	if err != nil {
		return n, err
	}

	return n, nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
