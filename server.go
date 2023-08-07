package main

import (
	"log"
	"net/http"

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
	upgrader := websocket.Upgrader{}
	lslogging.Configure(1, nil)
	r.GET("/ls", func(c *gin.Context) {
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
	})

	// run the server
	log.Fatal(r.Run(":8080"))
}
