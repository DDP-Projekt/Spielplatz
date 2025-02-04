package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/DDP-Projekt/DDPLS/ddpls"
	executables "github.com/DDP-Projekt/Spielplatz/server/execs_manager"
	"github.com/DDP-Projekt/Spielplatz/server/kddp"
	wsrw "github.com/DDP-Projekt/Spielplatz/server/websocket_rw"
	gin_pprof "github.com/gin-contrib/pprof"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
	"github.com/spf13/viper"
	lslogging "github.com/tliron/commonlog"
)

var DDPVERSION = "undefined"

func setup_logger(level slog.Level) {
	const time_fmt = time.DateTime + ".000"
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			TimeFormat: time_fmt,
			NoColor:    !isatty.IsTerminal(os.Stderr.Fd()),
			Level:      level,
		}),
	))
}

func fatal(msg string, args ...any) {
	// see https://pkg.go.dev/log/slog#example-package-Wrapping
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, fatal]
	r := slog.NewRecord(time.Now(), slog.LevelInfo+4, msg, pcs[0])
	r.Add(args...)

	slog.Default().Handler().Handle(context.Background(), r)
	panic(fmt.Errorf(msg))
}

func getLogger(c *gin.Context) *slog.Logger {
	if logger, ok := c.Get("logger"); ok {
		return logger.(*slog.Logger)
	}
	slog.Error("no logger in gin context")
	return slog.Default()
}

func setup_config() {
	viper.SetDefault("exe_cache_duration", time.Second*60)
	viper.SetDefault("run_timeout", time.Second*60)
	viper.SetDefault("port", "8080")
	viper.SetDefault("memory_limit_bytes", 4*(2<<29)) // 4 GiB
	viper.SetDefault("cpu_limit_percent", 50)
	viper.SetDefault("max_concurrent_processes", 50)
	viper.SetDefault("process_aquire_timeout", time.Second*3)
	viper.SetDefault("useHTTPS", false)
	viper.SetDefault("certPath", "")
	viper.SetDefault("keyPath", "")
	viper.SetDefault("pprof", false)
	viper.SetDefault("log_level", "INFO")
	viper.SetDefault("max_source_code_log_length", 100)

	var level slog.Level
	if err := level.UnmarshalText(
		[]byte(viper.GetString("log_level")),
	); err != nil {
		panic(err)
	}
	setup_logger(level)

	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		fatal("Error reading config file", "err", err)
	}

	settings := viper.AllSettings()
	settings_log := make([]any, 0, len(settings))
	for k, v := range settings {
		settings_log = append(settings_log, k, v)
	}
	slog.Info("configuration done", settings_log...)
}

func main() {
	setup_logger(slog.LevelInfo)
	setup_config()

	if err := kddp.InitializeSemaphore(viper.GetInt64("max_concurrent_processes")); err != nil {
		fatal("failed to initialize semaphore", "err", err)
	}

	r := gin.New()
	r.Use(
		gin.Recovery(),
		requestid.New(),
		func(c *gin.Context) {
			c.Set("logger",
				slog.Default().
					With("X-Request-ID", requestid.Get(c)).
					With("ip", c.ClientIP()),
			)
		},
	)

	// load html files as template
	r.LoadHTMLGlob("static/html/*")

	// serve node_modules/monaco-editor as /monaco-editor
	r.StaticFS("/monaco-editor", http.Dir("node_modules/monaco-editor"))
	// serve the static folder
	r.StaticFS("/static", http.Dir("static"))
	r.StaticFS("/img", http.Dir("img"))

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", DDPVERSION)
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

	r.GET("/health", serve_health)
	r.HEAD("/health", serve_health)

	if viper.GetString("pprof") != "" {
		gin_pprof.Register(r, "/debug/pprof")
	}

	// run the server

	var (
		use_https = viper.GetBool("useHTTPS")
		cert_path = viper.GetString("certPath")
		key_path  = viper.GetString("keyPath")
		port      = viper.GetString("port")
	)
	slog.Info("starting server", "https", use_https)
	if use_https {
		if cert_path == "" || key_path == "" {
			fatal("certPath and keyPath can not be empty!", "certPath", cert_path, "keyPath", key_path)
		}
		if err := r.RunTLS(":"+port, cert_path, key_path); err != nil {
			fatal("failed to run server", "err", err)
		}
	} else {
		if err := r.Run(":" + port); err != nil {
			fatal("failed to run server", "err", err)
		}
	}
}

var upgrader = websocket.Upgrader{}

// serves the /ls endpoint
func serve_ls(c *gin.Context) {
	logger := getLogger(c)
	logger.Info("new language server connection")
	// upgrade the connection to a websocket connection
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("failed to initialize websocket connection")
		return
	}
	defer ws.Close()

	ls := ddpls.NewDDPLS()
	ls.Server.ServeWebSocket(ws)
	logger.Info("language server connection closed")
}

// serves the /compile endpoint
func serve_compile(c *gin.Context) {
	logger := getLogger(c)
	type CompileRequest struct {
		Src string `json:"src"`
	}

	logger.Info("got compilation request")
	token, exe_path := executables.GenerateExeToken()
	logger = logger.With("token", token)
	logger.Info("generated token")
	// read the src json property from the request body
	var req CompileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("unmarshaling request", "err", err)
		executables.Delete(token)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	src_code := bytes.NewBufferString(req.Src)
	logger.Info("compiling the program", "source-code", truncString(req.Src, viper.GetInt("max_source_code_log_length")))
	// compile the program
	result, exe_path, err := kddp.CompileDDPProgram(src_code, token, exe_path, logger)
	if err != nil {
		logger.Error("compiling program", "err", err)
		executables.Delete(token)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	executables.Set(token, exe_path)

	logger.Info("compilation finished")
	// delete the executable after 3 minutes
	go func() {
		dur := viper.GetDuration("exe_cache_duration")
		time.Sleep(dur)
		if _, ok := executables.Get(token); ok {
			logger.Info("executable was unused for cache  duration, deleting it",
				"exe_path", exe_path,
				"cache_curation", dur,
				"token", token,
			)
			executables.RemoveExecutableFile(token, exe_path)
		}
	}()
	// send the result to the client
	c.JSON(http.StatusOK, result)
}

// serves the /run endpoint
func serve_run(c *gin.Context) {
	logger := getLogger(c)
	logger.Info("new run request")
	// upgrade the connection to a websocket connection
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("failed to initialize websocket connection")
		return
	}
	defer ws.Close()
	// get the token from the query
	token_str, ok := c.GetQuery("token")
	if !ok {
		logger.Warn("missing token in run request")
		// send a close message to the client with error
		ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInvalidFramePayloadData, "invalid token"))
		return
	}
	logger.Info("got run request token", "token", token_str)
	// parse token_str to uint64
	ti, err := strconv.ParseInt(token_str, 10, 64)
	token := executables.TokenType(ti)
	if err != nil {
		logger.Warn("invalid token")
		// send a close message to the client with error
		ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInvalidFramePayloadData, "invalid token"))
		return
	}
	logger = logger.With("token", token)
	// get the executable path from the executables map
	exe_path, ok := executables.Get(token)
	if !ok {
		logger.Warn("token was invalid")
		// send a close message to the client with error
		ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInvalidFramePayloadData, "invalid token"))
		return
	}
	logger = logger.With("exe_path", exe_path)

	args, _ := c.GetQueryArray("args")
	websocket_rw := wsrw.NewWebsocketRW(ws)
	// run the executable
	defer executables.RemoveExecutableFile(token, exe_path)

	logger.Info("running executable", "args", args)
	exitStatus, err := kddp.RunExecutable(exe_path, websocket_rw, websocket_rw.StdoutWriter(), websocket_rw.StderrWriter(), args, logger)
	if err != nil {
		logger.Error("failed to run executable", "err", err)
		websocket_rw.Close()
		// report error to client
		ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, err.Error()))
		return
	}
	logger.Info("executable ran successfully")
	websocket_rw.Close()
	ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, fmt.Sprintf("Das Programm wurde mit Code %d beendet", exitStatus)))
}

func truncString(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
