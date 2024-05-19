package execsmanager

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"time"
)

func fatal(msg string, args ...any) {
	// see https://pkg.go.dev/log/slog#example-package-Wrapping
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, fatal]
	r := slog.NewRecord(time.Now(), slog.LevelInfo+4, msg, pcs[0])
	r.Add(args...)

	slog.Default().Handler().Handle(context.Background(), r)
	panic(fmt.Errorf(msg))
}

const Exe_Dir = "playground_executables"

func init() {
	if _, err := os.Stat(Exe_Dir); err == nil {
		if err := os.RemoveAll(Exe_Dir); err != nil {
			fatal("could not delete directory for executables", "err", err, "dir", Exe_Dir)
		}
	}
	if err := os.Mkdir(Exe_Dir, os.ModePerm); err != nil {
		fatal("could not create directory for executables", "err", err, "dir", Exe_Dir)
	}
}
