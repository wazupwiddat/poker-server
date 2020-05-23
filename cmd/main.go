package main

import (
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"github.com/wazupwiddat/poker-server/server/controllers"
)

func main() {
	pokerServer, err := controllers.NewPokerServer(nil)
	if err != nil {
		log.Panic(err)
	}
	pokerServer.Setup()
	go pokerServer.Serve()
	defer pokerServer.Close()

	r := gin.Default()

	store := cookie.NewStore([]byte("poker-secret"))
	r.Use(sessions.Sessions("poker-session", store))

	r.LoadHTMLGlob("templates/*")
	r.Static("/assets", "./assets")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "newtable.html", nil)
	})
	r.GET("/socket.io/", func(c *gin.Context) {
		pokerServer.ServeHTTP(c.Writer, c.Request)
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
