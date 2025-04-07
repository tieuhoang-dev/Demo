package controllers

import (
	"Truyen_BE/config"
	"Truyen_BE/models"
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func GetStories(c *gin.Context) {
	if config.MongoDB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "MongoDB chưa được kết nối"})
		return
	}

	storyCollection := config.MongoDB.Collection("Stories")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := storyCollection.Find(ctx, bson.M{})
	if err != nil {
		log.Printf("Lỗi truy vấn mongoDB: %v", err) // Ghi log lỗi chi tiết
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi truy vấn mongoDB"})
		return
	}

	var stories []models.Story
	if err := cursor.All(ctx, &stories); err != nil {
		log.Printf("Lỗi giải mã dữ liệu: %v", err) // Ghi log lỗi giải mã
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi giải mã dữ liệu"})
		return
	}

	c.JSON(http.StatusOK, stories)
}

func SearchStoriesByName(c *gin.Context) {
	// Kiểm tra kết nối MongoDB trước khi truy cập
	if config.MongoDB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "MongoDB chưa được kết nối"})
		return
	}

	// Khởi tạo collection trong hàm thay vì toàn cục
	storyCollection := config.MongoDB.Collection("Stories")

	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu query 'name'"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"title": bson.M{"$regex": name, "$options": "i"},
	}

	cursor, err := storyCollection.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tìm kiếm"})
		return
	}

	var stories []models.Story
	if err := cursor.All(ctx, &stories); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi giải mã kết quả"})
		return
	}

	if len(stories) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "Không tìm thấy truyện"})
		return
	}

	c.JSON(http.StatusOK, stories)
}
func InsertStory(c *gin.Context) {
	var newStory models.Story
	if err := c.ShouldBindJSON(&newStory); err != nil {
		c.JSON(400, gin.H{"error": "Lỗi dữ liệu đầu vào"})
		return
	}
	newStory.CreatedAt = time.Now()
	newStory.UpdatedAt = time.Now()
	storyCollection := config.MongoDB.Collection("Stories")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	insertResult, err := storyCollection.InsertOne(ctx, newStory)
	if err != nil {
		log.Printf("Lỗi khi chèn truyện vào MongoDB: %v", err)
		c.JSON(500, gin.H{"error": "Lỗi khi chèn truyện vào MongoDB"})
		return
	}
	c.JSON(200, gin.H{
		"message": "Truyện đã được thêm thành công",
		"id":      insertResult.InsertedID,
	})
}
