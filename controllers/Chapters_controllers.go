package controllers

import (
	"Truyen_BE/config"
	"Truyen_BE/models"
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func InsertChapter(c *gin.Context) {
	var Newchapter models.Chapter
	if err := c.ShouldBindJSON(&Newchapter); err != nil {
		c.JSON(400, gin.H{"error": "Lỗi dữ liệu đầu vào"})
		return
	}
	Newchapter.CreatedAt = time.Now()
	Newchapter.UpdatedAt = time.Now()
	chapterCollection := config.MongoDB.Collection("Chapters")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	insertResult, err := chapterCollection.InsertOne(ctx, Newchapter)
	if err != nil {
		log.Printf("Lỗi khi chèn chương vào MongoDB: %v", err)
		c.JSON(500, gin.H{"error": "Lỗi khi chèn chương vào MongoDB"})
		return
	}
	c.JSON(200, gin.H{
		"message": "Chương đã được thêm thành công",
		"id":      insertResult.InsertedID,
	})
}
