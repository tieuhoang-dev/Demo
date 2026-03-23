package controllers

import (
	"Truyen_BE/config"
	"Truyen_BE/models"
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
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

	chapterCount, err := chapterCollection.CountDocuments(ctx, bson.M{"story_id": newChapter.StoryID})
	if err != nil {
		log.Printf("❌ Lỗi khi đếm số chương: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể xác định số chương"})
		return
	}

	newChapter.ChapterNumber = int(chapterCount) + 1

	_, err = chapterCollection.InsertOne(ctx, newChapter)
	if err != nil {
		log.Printf("❌ Lỗi khi chèn chương: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể thêm chương"})
		return
	}

	_, _ = storyCollection.UpdateOne(ctx,
		bson.M{"_id": newChapter.StoryID},
		bson.M{
			"$inc": bson.M{"chapters_count": 1},
			"$set": bson.M{"updated_at": time.Now()},
		},
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
	_, _ = chapterCollection.UpdateOne(ctx, bson.M{"_id": chapter.ID}, bson.M{"$inc": bson.M{"view_count": 1}})
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

	c.JSON(http.StatusOK, gin.H{
		"chapter":  chapter,
		"previous": previousChapter,
		"next":     nextChapter,
	})
}
func GetNewestChapters(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	chapterCollection := config.MongoDB.Collection("Chapters")
	storyCollection := config.MongoDB.Collection("Stories")

	cursor, err := chapterCollection.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(5))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy danh sách chương"})
		return
	}
	defer cursor.Close(ctx)

	var chapters []models.Chapter
	if err = cursor.All(ctx, &chapters); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy danh sách chương"})
		return
	}

	for i := range chapters {
		var story models.Story
		err = storyCollection.FindOne(ctx, bson.M{"_id": chapters[i].StoryID}).Decode(&story)
		if err == nil {
			chapters[i].Title = story.Title 
		}
	}

	c.JSON(http.StatusOK, chapters)
}
func InsertComment(c *gin.Context) {
	var comment models.Comment
	if err := c.ShouldBindJSON(&comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lỗi dữ liệu đầu vào"})
		return
	}

	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Chưa đăng nhập"})
		return
	}
	comment.UserID = userIDVal.(primitive.ObjectID)

	comment.Content = strings.TrimSpace(comment.Content)
	if comment.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nội dung không được để trống"})
		return
	}
	if len(comment.Content) > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nội dung quá dài"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := config.MongoDB.Collection("Stories").FindOne(ctx, bson.M{"_id": comment.StoryID}).Err(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Truyện không tồn tại"})
		return
	}
	if err := config.MongoDB.Collection("Chapters").FindOne(ctx, bson.M{"_id": comment.ChapterID}).Err(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Chương không tồn tại"})
		return
	}

	
	comment.ID = primitive.NewObjectID()
	comment.CreatedAt = time.Now()
	comment.UpdatedAt = time.Now()

	_, err := config.MongoDB.Collection("Comments").InsertOne(ctx, comment)
	if err != nil {
		log.Printf("❌ Lỗi khi chèn bình luận: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể thêm bình luận"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "✅ Đã thêm bình luận",
		"comment": gin.H{
			"id":         comment.ID.Hex(),
			"content":    comment.Content,
			"chapter_id": comment.ChapterID.Hex(),
			"story_id":   comment.StoryID.Hex(),
			"user_id":    comment.UserID.Hex(),
			"created_at": comment.CreatedAt,
		},
	})
}
func GetCommentsByChapterID(c *gin.Context) {
	chapterIDStr := c.Param("chapter_id")
	chapterID, err := primitive.ObjectIDFromHex(chapterIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID chương không hợp lệ"})
		return
	}

	// Phân trang
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	skip := (page - 1) * limit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	commentCollection := config.MongoDB.Collection("Comments")

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: 1}}).
		SetSkip(int64(skip)).
		SetLimit(int64(limit))

	cursor, err := commentCollection.Find(ctx, bson.M{"chapter_id": chapterID}, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy danh sách bình luận"})
		return
	}
	defer cursor.Close(ctx)

	var comments []models.Comment
	if err = cursor.All(ctx, &comments); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi xử lý bình luận"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"chapter_id": chapterID.Hex(),
		"page":       page,
		"limit":      limit,
		"comments":   comments,
	})
}
