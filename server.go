package main

import (
	"log"
	"net/http"

	"github.com/DDP-Projekt/DDPLS/ddpls"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/tliron/kutil/logging"
)

func main() {
	r := gin.Default()

	// load index html as template
	r.LoadHTMLFiles("index.html")

	// serve node_modules/monaco-editor as /monaco-editor
	r.StaticFS("/monaco-editor", http.Dir("node_modules/monaco-editor"))
	// serve the static folder
	r.StaticFS("/static", http.Dir("static"))

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	upgrader := websocket.Upgrader{}
	logging.Configure(1, nil)
	// write a websocket endpoint
	r.GET("/ls", func(c *gin.Context) {
		log.Println("new connection")
		// upgrade the connection to a websocket connection
		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println(err)
			return
		}
		defer ws.Close()

		ls := ddpls.NewDDPLS()
		ls.Server.ServeWebSocket(ws)
		log.Println("connection closed")
	})

	// run the server
	log.Fatal(r.Run(":8080"))
}
