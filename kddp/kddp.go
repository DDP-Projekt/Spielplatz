package kddp

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/viper"
	"golang.org/x/exp/constraints"
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

// constraint that satisfies `json:"token,string"`
type tokenType interface {
	string | constraints.Float | constraints.Integer | bool
}

// CompilerResult is the result of a compilation
// and will be sent to the client
type ProgramResult[TokenType tokenType] struct {
	ReturnCode int       `json:"returnCode,string"`
	Stderr     string    `json:"stderr"`
	Stdout     string    `json:"stdout"`
	Error      *string   `json:"error"`        // null if no error occurred
	Token      TokenType `json:"token,string"` // the token that was passed to compileDDPProgram
}

// compiles a DDP program and returns the result of the compilation,
// the path to the executable,
// and an error if one occurred
func CompileDDPProgram[TokenType tokenType](src io.Reader, token TokenType) (ProgramResult[TokenType], string, error) {
	exe_path := filepath.Join(exe_dir, "Spielplatz_"+fmt.Sprint(token))
	if runtime.GOOS == "windows" {
		exe_path += ".exe"
	}

	cmd := exec.Command("kddp", "kompiliere", "-o", exe_path)
	cmd.Stdin = src
	stderr := &strings.Builder{}
	stdout := &strings.Builder{}
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

	return ProgramResult[TokenType]{
		ReturnCode: cmd.ProcessState.ExitCode(),
		Stderr:     stderr.String(),
		Stdout:     stdout.String(),
		Error:      err_string,
		Token:      token,
	}, exe_path, nil
}

// runs an executable and returns the result of the execution
func RunExecutable(exe_path string, stdin io.Reader, stdout, stderr io.Writer, args ...string) error {
	timeout_chan := time.After(viper.GetDuration("run_timeout"))

	cmd := exec.Command(exe_path, args...)
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	cmd.Stdin = stdin

	if err := cmd.Start(); err != nil {
		return err
	}
	done := make(chan error, 1)

	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-timeout_chan:
		log.Printf("process %s excceeded timeout\n", exe_path)
		return cmd.Process.Kill()
	case err := <-done:
		return err
	}
}
