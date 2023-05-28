package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"backend/ai"
	"backend/contracts"
	"backend/database"
	"backend/discord"
	"backend/routes"

	. "backend/config"
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

	ai.Init(Config.OpenAIToken)

	if err := database.Init(); err != nil {
		fmt.Println(err)
		return
	}

	if err := contracts.Init(); err != nil {
		fmt.Println(err)
		return
	}

	if err := discord.Init(); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Bot Joined")
	defer func() {
		err := discord.Close()
		fmt.Println(err)
	}()

	r := gin.Default()
	r.Use(CORS)
	r.POST("/nonce", routes.GetNonce)
	r.POST("/login", routes.Login)
	r.POST("/logged-in", routes.IsAuth, routes.LoggedIn)
	r.POST("/logout", routes.LogOut)
	r.POST("/available", routes.IsAuth, routes.AvailableStories)
	r.POST("/start", routes.IsAuth, routes.Start)
	r.POST("/respond", routes.IsAuth, routes.Respond)
	r.POST("/image", routes.IsAuth, routes.Image)

	if err := r.Run(":" + strconv.Itoa(Config.Port)); err != nil {
		fmt.Println(err)
		return
	}
}
