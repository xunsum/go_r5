package main

import (
	"github.com/gin-gonic/gin"
	"go_r5/main/routers"
	"log"
)

func main() {
	//tool_code_pices.RunCode()

	router := gin.Default()
	routers.InitRouter(router)

	err := router.Run(":9060")
	if err != nil {
		log.Printf("Having problem starting gin.Engine: %v \n", err)
		return
	}
}
