package kddp

import (
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"golang.org/x/exp/constraints"
)

func init() {
	if _, err := exec.LookPath("kddp"); err != nil {
		log.Fatalf("kddp not found: %s", err)
	}
	if _, ok := os.LookupEnv("DDPPATH"); !ok {
		log.Println("DDPPATH not set, kddp might not work correctly")
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
func CompileDDPProgram[TokenType tokenType](src io.Reader, token TokenType, exe_path string) (ProgramResult[TokenType], string, error) {
	cmd := exec.Command("kddp", "kompiliere", "-o", exe_path, "--main", "seccomp_main.o", "--gcc_optionen", "-lseccomp")
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
func RunExecutable(exe_path string, stdin io.Reader, stdout, stderr io.Writer, args ...string) (int, error) {
	timeout_chan := time.After(viper.GetDuration("run_timeout"))

	cmd := exec.Command(exe_path, args...)
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	stdin_pipe, err := cmd.StdinPipe()
	if err != nil {
		return -1, err
	}

	if err := cmd.Start(); err != nil {
		return -1, err
	}
	done := make(chan error, 1)

	go func() {
		done <- cmd.Wait()
	}()
	go func() {
		if _, err := io.Copy(stdin_pipe, stdin); err != nil && !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
			log.Printf("error copying stdin to process: %s\n", err)
		}
		stdin_pipe.Close()
	}()

	select {
	case <-timeout_chan:
		log.Printf("process %s exceeded timeout\n", exe_path)
		err := errors.New("Process exceeded timeout")
		if kerr := cmd.Process.Kill(); kerr != nil {
			err = errors.Join(err, kerr)
		}
		return -1, err
	case err := <-done:
		return cmd.ProcessState.ExitCode(), err
	}
}
