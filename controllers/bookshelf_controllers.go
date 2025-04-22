package controllers

import (
	"Truyen_BE/config"
	"Truyen_BE/models"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AddToBookshelf(c *gin.Context) {
	// 1. Lấy user_id từ context (middleware đã gán)
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy user trong context"})
		return
	}
	userID := userIDValue.(primitive.ObjectID)

	// 2. Nhận dữ liệu từ body
	var input struct {
		StoryID       string `json:"story_id"`
		LastChapterID string `json:"last_chapter_id,omitempty"`
	}
	if err := c.ShouldBindJSON(&input); err != nil || input.StoryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu hoặc sai định dạng story_id"})
		return
	}

	// 3. Chuyển story_id sang ObjectID
	storyObjectID, err := primitive.ObjectIDFromHex(input.StoryID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "story_id không hợp lệ"})
		return
	}

	// 4. Tạo đối tượng BookshelfItem
	item := models.BookshelfItem{
		UserID:    userID,
		StoryID:   storyObjectID,
		ChapterID: primitive.NilObjectID, // Mặc định là nil nếu không có chapter
		AddedAt:   time.Now(),
		UpdatedAt: time.Now(),
	}

	// 5. Nếu có chương đang đọc thì lưu lại
	if input.LastChapterID != "" {
		lastChapterObjID, err := primitive.ObjectIDFromHex(input.LastChapterID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "last_chapter_id không hợp lệ"})
			return
		}
		item.ChapterID = lastChapterObjID
	}

	// 6. Kiểm tra nếu đã có trong tủ sách rồi
	collection := config.MongoDB.Collection("Bookshelf")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var existing models.BookshelfItem
	err = collection.FindOne(ctx, bson.M{
		"user_id":  userID,
		"story_id": storyObjectID,
	}).Decode(&existing)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Truyện này đã có trong tủ sách"})
		return
	}

	// 7. Thêm mới
	_, err = collection.InsertOne(ctx, item)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể thêm vào tủ sách"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "✅ Đã thêm truyện vào tủ sách"})
}
