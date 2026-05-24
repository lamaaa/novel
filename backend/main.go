package main

import (
	"fmt"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"novel-service/config"
	"novel-service/database"
	mcpServer "novel-service/mcp"
	"novel-service/router"
)

func main() {
	cfg := config.Load()
	database.Init(cfg)
	defer database.Close()

	r := gin.Default()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// 静态文件 - 前端
	r.Static("/assets", "../frontend/assets")
	r.StaticFile("/", "../frontend/index.html")
	r.NoRoute(func(c *gin.Context) {
		c.File("../frontend/index.html")
	})

	router.Setup(r)

	// MCP Streamable HTTP endpoint
	mcpServer.SetupRoutes(r)

	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	log.Printf("MCP endpoint: http://localhost%s/mcp", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
