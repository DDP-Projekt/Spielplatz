package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

const exe_dir = "playground_executables"

func init() {
	if _, err := exec.LookPath("kddp"); err != nil {
		log.Fatalf("kddp not found: %s", err)
	}
	if _, ok := os.LookupEnv("DDPPATH"); !ok {
		log.Println("DDPPATH not set, kddp might not work correctly")
	}
	if _, err := os.Stat(exe_dir); err == nil {
		if err := os.RemoveAll(exe_dir); err != nil {
			log.Fatalf("could not delete directory for executables: %s", err)
		}
	}
	if err := os.Mkdir(exe_dir, os.ModePerm); err != nil {
		log.Fatalf("could not create directory for executables: %s", err)
	}
}

type TokenType int64

// CompilerResult is the result of a compilation
// and will be sent to the client
type ProgramResult struct {
	ReturnCode int       `json:"returnCode,string"`
	Stderr     []byte    `json:"stderr"`
	Stdout     []byte    `json:"stdout"`
	Error      *string   `json:"error"`        // null if no error occurred
	Token      TokenType `json:"token,string"` // the token that was passed to compileDDPProgram
}

// compiles a DDP program and returns the result of the compilation,
// the path to the executable,
// and an error if one occurred
func compileDDPProgram(src io.Reader, token TokenType) (ProgramResult, string, error) {
	exe_path := filepath.Join(exe_dir, "Spielplatz_"+fmt.Sprint(token))
	if runtime.GOOS == "windows" {
		exe_path += ".exe"
	}

	cmd := exec.Command("kddp", "kompiliere", "-o", exe_path)
	cmd.Stdin = src
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	cmd.Stderr = stderr
	cmd.Stdout = stdout

	var err_string *string
	if err := cmd.Run(); err != nil {
		// delete exe_path if it exists
		if _, err := os.Stat(exe_path); err == nil {
			if err := os.Remove(exe_path); err != nil {
				log.Printf("could not delete executable: %s\n", err)
			}
		}
		err_string = new(string)
		*err_string = err.Error()
	}

	return ProgramResult{
		ReturnCode: cmd.ProcessState.ExitCode(),
		Stderr:     stderr.Bytes(),
		Stdout:     stdout.Bytes(),
		Error:      err_string,
		Token:      token,
	}, exe_path, nil
}

// runs an executable and returns the result of the execution
func runExecutable(exe_path string, stdin io.Reader, stdout, stderr io.Writer, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, exe_path, args...)
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	cmd.Stdin = stdin

	return cmd.Run()
}
