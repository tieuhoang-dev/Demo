package main

import (
	"Truyen_BE/config"
	"Truyen_BE/routes"
	"log"

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

	// Khởi tạo Gin
	r := gin.Default()
	// Cấu hình CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Cho phép mọi domain, tạm thời
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	// Gắn các routes
	routes.StoryRoutes(r)
	routes.ChapterRoutes(r)
	routes.UserRoutes(r)
	routes.BookshelfRoutes(r)
	// Chạy server tại cổng PORT từ .env
	if err := r.Run(":8080"); err != nil {
		log.Fatal("❌ Lỗi khi chạy server:", err)
	}
}
