package execsmanager

import (
	"log"
	"os"
)

const Exe_Dir = "playground_executables"

func init() {
	if _, err := os.Stat(Exe_Dir); err == nil {
		if err := os.RemoveAll(Exe_Dir); err != nil {
			log.Fatalf("could not delete directory for executables: %s", err)
		}
	}
	if err := os.Mkdir(Exe_Dir, os.ModePerm); err != nil {
		log.Fatalf("could not create directory for executables: %s", err)
	}
}
