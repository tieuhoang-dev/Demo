package controllers

import (
	"Truyen_BE/config"
	"Truyen_BE/models"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ChapterInput struct {
	StoryID string `json:"story_id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

// POST /chapters
func InsertChapter(c *gin.Context) {
	var newChapter models.Chapter
	if err := c.ShouldBindJSON(&newChapter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lỗi dữ liệu đầu vào"})
		return
	}

	if newChapter.StoryID.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu story_id"})
		return
	}

	newChapter.ID = primitive.NewObjectID()
	newChapter.CreatedAt = time.Now()
	newChapter.UpdatedAt = time.Now()
	newChapter.ViewCount = 0

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	chapterCollection := config.MongoDB.Collection("Chapters")
	storyCollection := config.MongoDB.Collection("Stories")

	// 🔢 Lấy số chương hiện có trong truyện
	chapterCount, err := chapterCollection.CountDocuments(ctx, bson.M{"story_id": newChapter.StoryID})
	if err != nil {
		log.Printf("❌ Lỗi khi đếm số chương: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể xác định số chương"})
		return
	}

	// Gán số chương mới
	newChapter.ChapterNumber = int(chapterCount) + 1

	// Insert chương
	_, err = chapterCollection.InsertOne(ctx, newChapter)
	if err != nil {
		log.Printf("❌ Lỗi khi chèn chương: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể thêm chương"})
		return
	}

	// Tăng chapters_count cho truyện tương ứng
	_, _ = storyCollection.UpdateOne(ctx,
		bson.M{"_id": newChapter.StoryID},
		bson.M{"$inc": bson.M{"chapters_count": 1}},
	)

	c.JSON(http.StatusOK, gin.H{
		"message":        "✅ Đã thêm chương mới",
		"id":             newChapter.ID.Hex(),
		"chapter_number": newChapter.ChapterNumber,
	})
}

// PUT /chapters/:id
func UpdateChapter(c *gin.Context) {
	chapterIDStr := c.Param("id")
	chapterID, err := primitive.ObjectIDFromHex(chapterIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID không hợp lệ"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lỗi dữ liệu đầu vào"})
		return
	}
	updates["updated_at"] = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	chapterCollection := config.MongoDB.Collection("Chapters")
	result, err := chapterCollection.UpdateOne(ctx,
		bson.M{"_id": chapterID},
		bson.M{"$set": updates},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể cập nhật chương"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "Không tìm thấy chương"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "✅ Đã cập nhật chương"})
}

// DELETE /chapters/:id
func DeleteChapter(c *gin.Context) {
	chapterIDStr := c.Param("id")
	chapterID, err := primitive.ObjectIDFromHex(chapterIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID không hợp lệ"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	chapterCollection := config.MongoDB.Collection("Chapters")
	storyCollection := config.MongoDB.Collection("Stories")

	// Tìm chương để lấy story_id
	var chapter models.Chapter
	err = chapterCollection.FindOne(ctx, bson.M{"_id": chapterID}).Decode(&chapter)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy chương"})
		return
	}

	// Xoá chương
	result, err := chapterCollection.DeleteOne(ctx, bson.M{"_id": chapterID})
	if err != nil || result.DeletedCount == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể xoá chương"})
		return
	}

	// Giảm chapters_count
	_, _ = storyCollection.UpdateOne(ctx,
		bson.M{"_id": chapter.StoryID},
		bson.M{"$inc": bson.M{"chapters_count": -1}},
	)

	c.JSON(http.StatusOK, gin.H{"message": "✅ Đã xoá chương"})
}
func GetChapterByStoryAndNumber(c *gin.Context) {
	storyID, err := primitive.ObjectIDFromHex(c.Param("story_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID truyện không hợp lệ"})
		return
	}

	number := c.Param("number")
	var chapterNumber int
	_, err = fmt.Sscanf(number, "%d", &chapterNumber)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Số chương không hợp lệ"})
		return
	}

	chapterCollection := config.MongoDB.Collection("Chapters")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var chapter models.Chapter
	err = chapterCollection.FindOne(ctx, bson.M{
		"story_id":       storyID,
		"chapter_number": chapterNumber,
		"is_Banned":      false,
		"is_hidden":      false,
	}).Decode(&chapter)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy chương hoặc chương đã bị ẩn/bị cấm"})
		return
	}
	//tăng view cho chương
	_, _ = chapterCollection.UpdateOne(ctx, bson.M{"_id": chapter.ID}, bson.M{"$inc": bson.M{"view_count": 1}})
	// Tăng view_count cho truyện tương ứng
	storyCollection := config.MongoDB.Collection("Stories")
	_, _ = storyCollection.UpdateOne(ctx,
		bson.M{"_id": chapter.StoryID},
		bson.M{"$inc": bson.M{"view_count": 1}},
	)
	c.JSON(http.StatusOK, chapter)
}

// GET /chapters/id/:id
func GetChapterByID(c *gin.Context) {
	chapterIDHex := c.Param("id")
	chapterID, err := primitive.ObjectIDFromHex(chapterIDHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID chương không hợp lệ"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	chapterCollection := config.MongoDB.Collection("Chapters")
	storyCollection := config.MongoDB.Collection("Stories")

	// Tăng view_count và lấy chương hiện tại
	var chapter models.Chapter
	err = chapterCollection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": chapterID},
		bson.M{"$inc": bson.M{"view_count": 1}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&chapter)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy chương"})
		return
	}

	// Tăng view_count cho truyện
	_, err = storyCollection.UpdateOne(
		ctx,
		bson.M{"_id": chapter.StoryID},
		bson.M{"$inc": bson.M{"view_count": 1}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi cập nhật lượt xem cho truyện"})
		return
	}

	// Tìm chương trước
	var previousChapter *models.Chapter = nil
	previousFilter := bson.M{
		"story_id":       chapter.StoryID,
		"chapter_number": bson.M{"$lt": chapter.ChapterNumber},
	}
	prevOptions := options.FindOne().SetSort(bson.D{{Key: "chapter_number", Value: -1}})
	var tempPrev models.Chapter
	err = chapterCollection.FindOne(ctx, previousFilter, prevOptions).Decode(&tempPrev)
	if err == nil {
		previousChapter = &tempPrev
	}

	// Tìm chương sau
	var nextChapter *models.Chapter = nil
	nextFilter := bson.M{
		"story_id":       chapter.StoryID,
		"chapter_number": bson.M{"$gt": chapter.ChapterNumber},
	}
	nextOptions := options.FindOne().SetSort(bson.D{{Key: "chapter_number", Value: 1}})
	var tempNext models.Chapter
	err = chapterCollection.FindOne(ctx, nextFilter, nextOptions).Decode(&tempNext)
	if err == nil {
		nextChapter = &tempNext
	}

	// Trả về dữ liệu
	c.JSON(http.StatusOK, gin.H{
		"chapter":  chapter,
		"previous": previousChapter,
		"next":     nextChapter,
	})
}
