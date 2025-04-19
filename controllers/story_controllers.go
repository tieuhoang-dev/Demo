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
	"go.mongodb.org/mongo-driver/bson/primitive"
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
func UpdateStory(c *gin.Context) {
	if config.MongoDB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "MongoDB chưa được kết nối"})
		return
	}
	storyCollection := config.MongoDB.Collection("Stories")
	// Lấy name của truyện từ query parameter
	StoryName := c.Param("name")
	if StoryName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "thiếu tên truyện"})
		return
	}
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(400, gin.H{"error": "Lỗi dữ liệu đầu vào"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	filter := bson.M{"title": StoryName}
	update := bson.M{"$set": updates}
	result, err := storyCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể cập nhật truyện"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "Không tìm thấy truyện"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "✅ Đã cập nhật truyện"})
}
func DeleteStory(c *gin.Context) {
	if config.MongoDB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "MongoDB chưa được kết nối"})
		return
	}

	storyCollection := config.MongoDB.Collection("Stories")
	chapterCollection := config.MongoDB.Collection("Chapters")

	// Lấy name của truyện từ query parameter
	storyName := c.Param("name")
	if storyName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu tên truyện"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Trước tiên: tìm truyện để lấy _id
	var story struct {
		ID primitive.ObjectID `bson:"_id"`
	}
	err := storyCollection.FindOne(ctx, bson.M{"title": storyName}).Decode(&story)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy truyện"})
		return
	}

	// Xóa các chương liên quan
	_, err = chapterCollection.DeleteMany(ctx, bson.M{"StoryID": story.ID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể xóa chương liên quan"})
		return
	}

	// Sau đó xóa truyện
	result, err := storyCollection.DeleteOne(ctx, bson.M{"_id": story.ID})
	if err != nil || result.DeletedCount == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể xóa truyện"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "✅ Xóa truyện và chương thành công"})
}

func GetChaptersByStoryName(c *gin.Context) {
	if config.MongoDB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "MongoDB chưa được kết nối"})
		return
	}

	storyCollection := config.MongoDB.Collection("Stories")
	chapterCollection := config.MongoDB.Collection("Chapters")

	// Lấy name của truyện từ query parameter
	storyName := c.Query("name")
	if storyName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "thiếu tên truyện"})
		return
	}

	// Tìm truyện theo tên
	var story struct {
		ID primitive.ObjectID `bson:"_id"`
	}
	err := storyCollection.FindOne(c, bson.M{"title": storyName}).Decode(&story)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy truyện"})
		return
	}

	// Tìm các chương theo storyId
	cursor, err := chapterCollection.Find(c, bson.M{"StoryID": story.ID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể truy vấn chương"})
		return
	}
	defer cursor.Close(c)

	// Load danh sách chương
	var chapters []bson.M
	if err = cursor.All(c, &chapters); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi đọc dữ liệu chương"})
		return
	}

	// Trả về JSON
	c.JSON(http.StatusOK, gin.H{
		"story_id": story.ID.Hex(),
		"chapters": chapters,
	})
}
