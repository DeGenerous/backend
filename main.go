package main

import (
	"backend/ai"
	"backend/routes"
	"fmt"
	"strings"

	. "backend/config"
	"github.com/gin-gonic/gin"
)

func CORS(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", Config.CORSUrls[0])
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Headers",
		"Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

	for _, url := range Config.CORSUrls {
		origin := c.Request.Header.Get("Origin")
		if strings.Contains(origin, url) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}
	}

	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(204)
		return
	}

	c.Next()
}

func main() {
	if err := Config.Read(); err != nil {
		fmt.Println(err)
		return
	}

	ai.Init(Config.Token)

	r := gin.Default()
	r.Use(CORS)
	r.POST("/start", routes.Start)
	r.POST("/respond", routes.Respond)

	if err := r.Run(":3000"); err != nil {
		fmt.Println(err)
		return
	}
}
