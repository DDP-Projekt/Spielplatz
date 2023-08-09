package main

import (
	"bytes"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/DDP-Projekt/DDPLS/ddpls"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	lslogging "github.com/tliron/kutil/logging"
)

func main() {
	r := gin.Default()

	// load index html as template
	r.LoadHTMLFiles("index.html")

	// serve node_modules/monaco-editor as /monaco-editor
	r.StaticFS("/monaco-editor", http.Dir("node_modules/monaco-editor"))
	// serve the static folder
	r.StaticFS("/static", http.Dir("static"))
	r.StaticFS("/img", http.Dir("img"))

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// websocket endpoint to connect to the language server
	lslogging.Configure(1, nil)
	r.GET("/ls", serve_ls)

	// endpoint to compile a ddp program
	r.POST("/compile", serve_compile)
	r.GET("/run", serve_run)

	// run the server
	log.Fatal(r.Run(":8080"))
}

var upgrader = websocket.Upgrader{}

// serves the /ls endpoint
func serve_ls(c *gin.Context) {
	log.Printf("new connection to %s\n", c.ClientIP())
	// upgrade the connection to a websocket connection
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer ws.Close()

	ls := ddpls.NewDDPLS()
	ls.Server.ServeWebSocket(ws)
	log.Printf("connection with %s closed\n", c.ClientIP())
}

var tokenGenerator = rand.NewSource(time.Now().UnixNano())
var executables = NewSyncMap[TokenType, string]()

func deleteExecutable(token TokenType, exe_path string) {
	log.Printf("deleting %s\n", exe_path)
	if _, ok := executables.Get(token); ok {
		if err := os.Remove(exe_path); err != nil {
			log.Printf("could not delete executable: %s\n", err)
		}
	}
	executables.Delete(token)
}

// generates a token and adds it to the executables map
func generateExeToken() TokenType {
	for {
		tok := TokenType(tokenGenerator.Int63())
		if _, ok := executables.Get(tok); !ok {
			executables.Set(tok, "")
			return tok
		}
	}
}

type CompileRequest struct {
	Src string `json:"src"`
}

// serves the /compile endpoint
func serve_compile(c *gin.Context) {
	log.Printf("new compilation request from %s\n", c.ClientIP())
	token := generateExeToken()
	// read the src json property from the request body
	var req CompileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println(err)
		executables.Delete(token)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	src_code := bytes.NewBufferString(req.Src)
	// compile the program
	result, exe_path, err := compileDDPProgram(src_code, token)
	if err != nil {
		log.Println(err)
		executables.Delete(token)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	executables.Set(token, exe_path)

	log.Printf("compilation of program %d finished\n", token)
	// delete the executable after 3 minutes
	go func() {
		time.Sleep(3 * time.Minute)
		deleteExecutable(token, exe_path)
	}()
	// send the result to the client
	c.JSON(http.StatusOK, result)
}

// serves the /run endpoint
func serve_run(c *gin.Context) {
	log.Printf("new run request from %s\n", c.ClientIP())
	// upgrade the connection to a websocket connection
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer ws.Close()
	// get the token from the query
	token_str, ok := c.GetQuery("token")
	if !ok {
		// send a close message to the client with error
		ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInvalidFramePayloadData, "invalid token"))
		return
	}
	// parse token_str to uint64
	ti, err := strconv.ParseInt(token_str, 10, 64)
	token := TokenType(ti)
	if err != nil {
		// send a close message to the client with error
		ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInvalidFramePayloadData, "invalid token"))
		return
	}
	// get the executable path from the executables map
	exe_path, ok := executables.Get(token)
	if !ok {
		log.Printf("client requested run for invalid token %d", token)
		// send a close message to the client with error
		ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInvalidFramePayloadData, "invalid token"))
		return
	}
	args, _ := c.GetQueryArray("args")
	websocket_rw := NewWebsocketRW(ws)
	// run the executable
	defer deleteExecutable(token, exe_path)
	if err := runExecutable(exe_path, websocket_rw, websocket_rw, websocket_rw, args...); err != nil {
		log.Println(err)
		websocket_rw.Close()
		// report error to client
		ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, err.Error()))
		return
	}
	websocket_rw.Close()
	ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}
