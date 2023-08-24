package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/DDP-Projekt/DDPLS/ddpls"
	executables "github.com/DDP-Projekt/Spielplatz/server/execs_manager"
	"github.com/DDP-Projekt/Spielplatz/server/kddp"
	wsrw "github.com/DDP-Projekt/Spielplatz/server/websocket_rw"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	lslogging "github.com/tliron/kutil/logging"
)

func setup_config() {
	viper.SetDefault("exe_cache_duration", time.Second*60)
	viper.SetDefault("run_timeout", time.Second*60)
	viper.SetDefault("port", "8080")

	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %s\n", err)
	}
}

func main() {
	setup_config()

	r := gin.Default()

	// load html files as template
	r.LoadHTMLGlob("static/html/*")

	// serve node_modules/monaco-editor as /monaco-editor
	r.StaticFS("/monaco-editor", http.Dir("node_modules/monaco-editor"))
	// serve the static folder
	r.StaticFS("/static", http.Dir("static"))
	r.StaticFS("/img", http.Dir("img"))

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.GET("/embed", func(c *gin.Context) {
		c.HTML(http.StatusOK, "embed.html", nil)
	})

	// websocket endpoint to connect to the language server
	lslogging.Configure(1, nil)
	r.GET("/ls", serve_ls)

	// endpoint to compile a ddp program
	r.POST("/compile", serve_compile)
	r.GET("/run", serve_run)

	// run the server
	log.Fatal(r.Run(":" + viper.GetString("port")))
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
	token, exe_path := executables.GenerateExeToken()
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
	result, exe_path, err := kddp.CompileDDPProgram(src_code, token, exe_path)
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
		dur := viper.GetDuration("exe_cache_duration")
		time.Sleep(dur)
		if _, ok := executables.Get(token); ok {
			log.Printf("executable %s was unused for %s, deleting it", exe_path, dur)
			executables.RemoveExecutableFile(token, exe_path)
		}
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
	exitStatus, err := kddp.RunExecutable(exe_path, websocket_rw, websocket_rw.StdoutWriter(), websocket_rw.StderrWriter(), args...)
	if err != nil {
		log.Println(err)
		websocket_rw.Close()
		// report error to client
		ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, err.Error()))
		return
	}
	websocket_rw.Close()
	ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, fmt.Sprintf("Process exited with status %d", exitStatus)))
}
