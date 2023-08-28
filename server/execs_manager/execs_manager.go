/*
package excsmanager manages the executable files that are created by the server
*/
package execsmanager

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/DDP-Projekt/Spielplatz/server/execs_manager/syncmap"
)

type TokenType int64

var tokenGenerator = rand.NewSource(time.Now().UnixNano())
var executables = syncmap.NewSyncMap[TokenType, string]()

func Get(token TokenType) (string, bool) {
	return executables.Get(token)
}

func Set(token TokenType, exe_path string) {
	executables.Set(token, exe_path)
}

func Delete(token TokenType) {
	executables.Delete(token)
}

func RemoveExecutableFile(token TokenType, exe_path string) {
	if _, ok := executables.Get(token); ok {
		log.Printf("deleting %s\n", exe_path)
		if err := os.Remove(exe_path); err != nil {
			log.Printf("could not delete executable: %s\n", err)
		}
		executables.Delete(token)
	}
}

// generates a token and adds it to the executables map
// returns the token and the path to the executable
func GenerateExeToken() (TokenType, string) {
	for {
		tok := TokenType(tokenGenerator.Int63())
		if _, ok := executables.Get(tok); !ok {
			executables.Set(tok, "")
			return tok, genExePath(tok)
		}
	}
}

func genExePath(token TokenType) string {
	exe_path := filepath.Join(Exe_Dir, "Spielplatz_"+fmt.Sprint(token))
	if runtime.GOOS == "windows" {
		exe_path += ".exe"
	}
	return exe_path
}
