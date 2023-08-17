package kddp

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/criyle/go-sandbox/pkg/forkexec"
	"github.com/criyle/go-sandbox/pkg/seccomp"
	bseccomp "github.com/elastic/go-seccomp-bpf"
	"golang.org/x/exp/constraints"
	"golang.org/x/net/bpf"
	"golang.org/x/sys/unix"
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
func RunExecutable(exe_path string, stdin io.Reader, stdout, stderr io.Writer, args ...string) (int, error) {
	// timeout_chan := time.After(viper.GetDuration("run_timeout"))

	// create the seccomp filter
	seccomp_filter, err := createSeccompFilter()
	if err != nil {
		return 1, err
	}

	// open the executable
	bin, err := os.Open(exe_path)
	if err != nil {
		return 1, err
	}
	defer bin.Close()

	// create the stdin/stdout/stderr pipes
	std_pipes, err := createPipes()
	if err != nil {
		return 1, err
	}

	// create the command
	cmd := forkexec.Runner{
		Args:     args,
		ExecFile: bin.Fd(),
		Files:    []uintptr{std_pipes[0][1].Fd(), std_pipes[1][1].Fd(), std_pipes[2][1].Fd()},
		// RLimits: []syscall.Rlimit{} // TODO
		WorkDir: filepath.Dir(exe_path),
		Seccomp: seccomp_filter,
	}
	pid, err := cmd.Start()
	if err != nil {
		return 1, err
	}

	/*
		// at the end
		defer func(pipes [3][2]*os.File) {
			var err error
			for _, pipe := range pipes {
				if cerr := pipe[1].Close(); cerr != nil {
					err = errors.Join(err, cerr)
				}
			}
			if err != nil {
				log.Println(err)
			}
		}(std_pipes)*/

	// close our pipe ends to not block indefinetly
	for _, pipe := range std_pipes {
		if cerr := pipe[1].Close(); cerr != nil {
			err = errors.Join(err, cerr)
		}
	}
	if err != nil {
		if kerr := unix.Kill(pid, syscall.SIGKILL); kerr != nil {
			err = errors.Join(err, kerr)
		}
		return 1, err
	}

	start_pipe_goroutines(std_pipes, stdin, stdout, stderr)

	// wait for the process to finish
	var ws unix.WaitStatus
	if _, err = unix.Wait4(pid, &ws, 0, nil); err != nil {
		return ws.ExitStatus(), err
	}

	// return the result
	if ws.Stopped() {
		return 0, fmt.Errorf("stopped: %s %d", ws.StopSignal(), ws.TrapCause())
	} else if ws.Signaled() {
		return -1, fmt.Errorf("process was terminated by signal %s", ws.Signal())
	}
	return ws.ExitStatus(), nil
}

func start_pipe_goroutines(std_pipes [3][2]*os.File, stdin io.Reader, stdout, stderr io.Writer) {
	go func(pipe *os.File, stdin io.Reader) {
		_, err := io.Copy(pipe, stdin)
		if cerr := pipe.Close(); cerr != nil {
			err = errors.Join(err, cerr)
		}
		if err != nil {
			log.Println(err)
		}
	}(std_pipes[0][0], stdin)
	go func(pipe *os.File, stdout io.Writer) {
		_, err := io.Copy(stdout, pipe)
		if cerr := pipe.Close(); cerr != nil {
			err = errors.Join(err, cerr)
		}
		if err != nil {
			log.Println(err)
		}
	}(std_pipes[1][0], stdout)
	go func(pipe *os.File, stderr io.Writer) {
		_, err := io.Copy(stderr, pipe)
		if cerr := pipe.Close(); cerr != nil {
			err = errors.Join(err, cerr)
		}
		if err != nil {
			log.Println(err)
		}
	}(std_pipes[2][0], stderr)
}

var allowed_syscalls = []string{
	"write",
	"read",
	"mmap",
	"execveat",
	"brk",
	"close",
	"mprotect",
	"munmap",
	"pread64",
	"arch_prctl",
	"set_tid_address",
	"openat",
	"newfstatat",
	"set_robust_list",
	"getrandom",
	"rseq",
	"exit_group",
	"exit",
}

func createPipes() ([3][2]*os.File, error) {
	close_pipe := func(p [2]*os.File) error {
		if p[0] == nil && p[1] == nil {
			return nil
		}
		err := p[0].Close()
		if cerr := p[1].Close(); cerr != nil {
			err = errors.Join(err, cerr)
		}
		return err
	}

	// open 3 os pipes and close them all in case of error
	var pipes [3][2]*os.File
	for i := 0; i < 3; i++ {
		r, w, err := os.Pipe()
		if i == 0 { // for stdin, we want the read end of the pipe
			r, w = w, r
		}
		// in case of error, close the already opened pipes
		if err != nil {
			for _, pipe := range pipes {
				if cerr := close_pipe(pipe); cerr != nil {
					err = errors.Join(err, cerr)
				}
			}
			return pipes, err
		}
		pipes[i] = [2]*os.File{r, w}
	}
	return pipes, nil
}

var seccomp_policy = bseccomp.Policy{
	DefaultAction: bseccomp.ActionErrno, // error if the syscall is not allowed
	Syscalls: []bseccomp.SyscallGroup{
		{
			Names:  allowed_syscalls,
			Action: bseccomp.ActionAllow,
		},
	},
}

func createSeccompFilter() (*syscall.SockFprog, error) {
	policy, err := seccomp_policy.Assemble()
	if err != nil {
		return nil, err
	}
	seccomp_filter, err := exportBPF(policy)
	if err != nil {
		return nil, err
	}
	return seccomp_filter.SockFprog(), nil
}

// ExportBPF convert libseccomp filter to kernel readable BPF content
func exportBPF(filter []bpf.Instruction) (seccomp.Filter, error) {
	raw, err := bpf.Assemble(filter)
	if err != nil {
		return nil, err
	}

	result := make([]syscall.SockFilter, 0, len(raw))
	for _, instruction := range raw {
		result = append(result, syscall.SockFilter{
			Code: instruction.Op,
			Jt:   instruction.Jt,
			Jf:   instruction.Jf,
			K:    instruction.K,
		})
	}
	return result, nil
}
