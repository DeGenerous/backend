package main

import (
	"backend/ai"
	"backend/routes"
	"fmt"

	. "backend/config"
	"github.com/gin-gonic/gin"
)

func main() {
	if err := Config.Read(); err != nil {
		fmt.Println(err)
		return
	}

	ai.Init(Config.Token)

	r := gin.Default()
	r.POST("/start", routes.Start)
	r.POST("/respond", routes.Respond)

	if err := r.Run(":8080"); err != nil {
		fmt.Println(err)
		return
	}
}
