package execsmanager

import (
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/DDP-Projekt/Spielplatz/syncmap"
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
		log.Printf("deleting %s after timeout\n", exe_path)
		if err := os.Remove(exe_path); err != nil {
			log.Printf("could not delete executable: %s\n", err)
		}
		executables.Delete(token)
	}
}

// generates a token and adds it to the executables map
func GenerateExeToken() TokenType {
	for {
		tok := TokenType(tokenGenerator.Int63())
		if _, ok := executables.Get(tok); !ok {
			executables.Set(tok, "")
			return tok
		}
	}
}
