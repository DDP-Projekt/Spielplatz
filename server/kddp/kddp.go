/*
package kddp manages the compilation and execution of the created programs
*/
package kddp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"golang.org/x/exp/constraints"
	"golang.org/x/sync/semaphore"
)

func init() {
	if _, err := exec.LookPath("kddp"); err != nil {
		log.Fatalf("kddp not found: %s", err)
	}
	if _, ok := os.LookupEnv("DDPPATH"); !ok {
		log.Println("DDPPATH not set, kddp might not work correctly")
	}
}

var proc_sem *semaphore.Weighted

func InitializeSemaphore(weight int64) error {
	if weight < 1 {
		return errors.New("weight must be at least 1")
	}
	proc_sem = semaphore.NewWeighted(weight)
	return nil
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
	cmd := exec.Command("kddp", "kompiliere", "-o", exe_path, "--main", "seccomp_main.o", "--gcc_optionen=-lseccomp -static -no-pie")
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
	if proc_sem != nil {
		sem_ctx, sem_cancel := context.WithTimeout(context.Background(), viper.GetDuration("process_aquire_timeout"))
		if err := proc_sem.Acquire(sem_ctx, 1); err != nil {
			sem_cancel()
			return -1, errors.Join(errors.New("Der Server ist momentan ausgelastet, versuchen sie es später erneut"), err)
		}
		defer proc_sem.Release(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("run_timeout"))
	defer cancel()

	var err error
	exe_path, err = filepath.Abs(exe_path)
	if err != nil {
		return -1, err
	}

	args = append([]string{exe_path}, args...)

	cmd := exec.CommandContext(ctx, "./seccomp_exec", args...)
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	stdin_pipe, err := cmd.StdinPipe()
	if err != nil {
		return -1, err
	}

	if err := cmd.Start(); err != nil {
		return -1, err
	}

	done := make(chan error)

	go func() {
		done <- cmd.Wait()
	}()

	go func() {
		if _, err := io.Copy(stdin_pipe, stdin); err != nil && !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
			log.Printf("error copying stdin to process: %s\n", err)
			cancel()
		}
		stdin_pipe.Close()
	}()

	err = <-done
	if cerr := ctx.Err(); cerr != nil {
		switch cerr {
		case context.DeadlineExceeded:
			err = errors.New("Das Programm hat die Frist überschritten")
		case context.Canceled:
			err = fmt.Errorf("Das Programm wurde abgebrochen: %s", cerr.Error())
		default:
			err = cerr
		}
	}
	return cmd.ProcessState.ExitCode(), err
}
