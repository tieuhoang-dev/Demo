package main

import (
	"Truyen_BE/config"
	"Truyen_BE/routes"
	"log"

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
