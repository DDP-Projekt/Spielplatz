/*
package kddp manages the compilation and execution of the created programs
*/
package kddp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"golang.org/x/exp/constraints"
	"golang.org/x/sync/semaphore"
)

func fatal(msg string, args ...any) {
	slog.Log(context.Background(), slog.LevelError+4, msg, args...)
	panic(fmt.Errorf(msg))
}

func init() {
	if _, err := exec.LookPath("kddp"); err != nil {
		fatal("kddp not found", "err", err)
	}
	if _, ok := os.LookupEnv("DDPPATH"); !ok {
		slog.Warn("DDPPATH not set, kddp might not work correctly")
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
func CompileDDPProgram[TokenType tokenType](src io.Reader, token TokenType, exe_path string, logger *slog.Logger) (ProgramResult[TokenType], string, error) {
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
				logger.Warn("failed to delete executable after error", "err", err)
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
func RunExecutable(exe_path string, stdin io.Reader, stdout, stderr io.Writer, args []string, logger *slog.Logger) (int, error) {
	if proc_sem != nil {
		sem_ctx, sem_cancel := context.WithTimeout(context.Background(), viper.GetDuration("process_aquire_timeout"))
		defer sem_cancel()
		if err := proc_sem.Acquire(sem_ctx, 1); err != nil {
			return -1, errors.Join(errors.New("Der Server ist momentan ausgelastet, versuchen sie es später erneut"), err)
		}
		defer proc_sem.Release(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("run_timeout"))
	defer cancel()

	var err error
	exe_path, err = filepath.Abs(exe_path)
	if err != nil {
		logger.Error("failed to get absolute path to executable", "err", err)
		return -1, fmt.Errorf("error getting absoulte path to executable: %w", err)
	}
	logger = logger.With("exe_path", exe_path)

	args = append([]string{exe_path}, args...)

	cmd := exec.CommandContext(ctx, "./seccomp_exec", args...)
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	stdin_pipe, err := cmd.StdinPipe()
	if err != nil {
		logger.Error("failed to create stdin pipe", "err", err)
		return -1, fmt.Errorf("error creating stdin pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		logger.Error("failed to start executable", "err", err)
		return -1, fmt.Errorf("error starting executable: %w", err)
	}

	done := make(chan error)
	is_done := atomic.Bool{}

	go func() {
		err := cmd.Wait()
		is_done.Store(true)
		done <- err
	}()

	go func() {
		isBadReadErr := func(err error) bool {
			if err == nil {
				return false
			}
			// ErrCloseSent is only a true error, if the program was still running when the error occured
			if errors.Is(err, websocket.ErrCloseSent) {
				return !is_done.Load()
			}
			return !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway)
		}

		if _, err := io.Copy(stdin_pipe, stdin); isBadReadErr(err) {
			logger.Warn("error copying stdin to process", "err", err)
			cancel()
		}
		logger.Info("closing stdin pipe")
		stdin_pipe.Close()
	}()

	err = <-done
	if cerr := ctx.Err(); cerr != nil {
		switch cerr {
		case context.DeadlineExceeded:
			logger.Info("deadline exceeded")
			err = errors.New("Das Programm hat die Frist überschritten")
		case context.Canceled:
			logger.Info("program cancelled")
			err = fmt.Errorf("Das Programm wurde abgebrochen: %w", cerr)
		default:
			err = cerr
		}
	}
	return cmd.ProcessState.ExitCode(), err
}

// CompilerResult is the result of a compilation
// and will be sent to the client
type VersionResult struct {
	ReturnCode int    `json:"returnCode"`
	Stdout     string `json:"stdout"`
}

func GetKDDPVersion() (VersionResult, error) {
	cmd := exec.Command("kddp", "version")
	stderr := &strings.Builder{}
	stdout := &strings.Builder{}
	cmd.Stderr = stderr
	cmd.Stdout = stdout

	err := cmd.Run()

	if err != nil && stderr.String() != "" {
		err = fmt.Errorf(stderr.String())
	}

	return VersionResult{
		ReturnCode: cmd.ProcessState.ExitCode(),
		Stdout:     stdout.String(),
	}, err
}
