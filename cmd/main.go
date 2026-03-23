package main

import (
	"Truyen_BE/config"
	"Truyen_BE/routes"
	"log"
	"os"

	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load các biến môi trường
	config.LoadEnv()

	// Kết nối MongoDB
	config.ConnectDB()

	// Kiểm tra kết nối MongoDB sau khi connect
	if config.MongoDB == nil {
		log.Fatal("❌ Kết nối MongoDB không thành công!")
	}
	var FE_URL = os.Getenv("FE_URL")
	// Khởi tạo Gin
	r := gin.Default()
	// Cấu hình CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{FE_URL},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Gắn các routes
	routes.StoryRoutes(r)
	routes.ChapterRoutes(r)
	routes.UserRoutes(r)
	routes.BookshelfRoutes(r)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
	log.Fatal(r.Run(port))
}
