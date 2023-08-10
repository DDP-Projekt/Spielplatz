package main

import (
	"bytes"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/DDP-Projekt/DDPLS/ddpls"
	executables "github.com/DDP-Projekt/Spielplatz/execs_manager"
	"github.com/DDP-Projekt/Spielplatz/kddp"
	wsrw "github.com/DDP-Projekt/Spielplatz/websocket_rw"
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

// serves the /compile endpoint
func serve_compile(c *gin.Context) {
	type CompileRequest struct {
		Src string `json:"src"`
	}

	log.Printf("new compilation request from %s\n", c.ClientIP())
	token := executables.GenerateExeToken()
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
	result, exe_path, err := kddp.CompileDDPProgram(src_code, token)
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
		executables.RemoveExecutableFile(token, exe_path)
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
	token := executables.TokenType(ti)
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
	websocket_rw := wsrw.NewWebsocketRW(ws)
	// run the executable
	defer executables.RemoveExecutableFile(token, exe_path)
	if err := kddp.RunExecutable(exe_path, websocket_rw, websocket_rw, websocket_rw, args...); err != nil {
		log.Println(err)
		websocket_rw.Close()
		// report error to client
		ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, err.Error()))
		return
	}
	websocket_rw.Close()
	ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}
