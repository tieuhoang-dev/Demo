package controllers

import (
	"Truyen_BE/config"
	"Truyen_BE/models"
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GET /stories
func GetStories(c *gin.Context) {
	storyCollection := config.MongoDB.Collection("Stories")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := storyCollection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi truy vấn MongoDB"})
		return
	}

	var stories []models.Story
	if err := cursor.All(ctx, &stories); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi giải mã dữ liệu"})
		return
	}

	c.JSON(http.StatusOK, stories)
}

// GET /stories/search?name=abc
func SearchStoriesByName(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu query 'name'"})
		return
	}

	storyCollection := config.MongoDB.Collection("Stories")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"title": bson.M{"$regex": name, "$options": "i"}}
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

	c.JSON(http.StatusOK, stories)
}

// POST /stories
func InsertStory(c *gin.Context) {
	var newStory models.Story
	if err := c.ShouldBindJSON(&newStory); err != nil {
		c.JSON(400, gin.H{"error": "Lỗi dữ liệu đầu vào"})
		return
	}

	newStory.ID = primitive.NewObjectID()
	newStory.CreatedAt = time.Now()
	newStory.UpdatedAt = time.Now()
	newStory.ChaptersCount = 0
	newStory.ViewCount = 0
	newStory.IsHidden = false
	newStory.IsFeatured = false

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	storyCollection := config.MongoDB.Collection("Stories")
	_, err := storyCollection.InsertOne(ctx, newStory)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi chèn truyện"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "✅ Truyện đã được thêm", "id": newStory.ID.Hex()})
}

// PUT /stories/:id
func UpdateStory(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(400, gin.H{"error": "ID không hợp lệ"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(400, gin.H{"error": "Dữ liệu đầu vào sai"})
		return
	}
	updates["updated_at"] = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	storyCollection := config.MongoDB.Collection("Stories")
	result, err := storyCollection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": updates})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể cập nhật truyện"})
		return
	}
	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "Không tìm thấy truyện"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "✅ Truyện đã được cập nhật"})
}

// DELETE /stories/:id
func DeleteStory(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(400, gin.H{"error": "ID không hợp lệ"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	storyCollection := config.MongoDB.Collection("Stories")
	chapterCollection := config.MongoDB.Collection("Chapters")

	// Xoá các chương trước
	_, err = chapterCollection.DeleteMany(ctx, bson.M{"story_id": objectID})
	if err != nil {
		c.JSON(500, gin.H{"error": "Không thể xoá các chương"})
		return
	}

	// Xoá truyện
	result, err := storyCollection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil || result.DeletedCount == 0 {
		c.JSON(500, gin.H{"error": "Không thể xoá truyện"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "✅ Đã xoá truyện và các chương"})
}

// GET /stories/:id/chapters
func GetChaptersByStoryID(c *gin.Context) {
	id := c.Param("id")
	storyID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(400, gin.H{"error": "ID truyện không hợp lệ"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	chapterCollection := config.MongoDB.Collection("Chapters")
	cursor, err := chapterCollection.Find(ctx, bson.M{"story_id": storyID})
	if err != nil {
		c.JSON(500, gin.H{"error": "Không thể truy vấn chương"})
		return
	}
	defer cursor.Close(ctx)

	var chapters []models.Chapter
	if err := cursor.All(ctx, &chapters); err != nil {
		c.JSON(500, gin.H{"error": "Lỗi đọc chương"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"story_id": storyID.Hex(),
		"chapters": chapters,
	})
}

// GET /stories/filter
func FilterStories(c *gin.Context) {
	storyCollection := config.MongoDB.Collection("Stories")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Params
	genre := c.Query("genre")
	status := c.Query("status")
	author := c.Query("author")
	sort := c.DefaultQuery("sort", "updated_desc")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	skip := int64((page - 1) * limit)
	limit64 := int64(limit)

	// Filter
	filter := bson.M{}
	if genre != "" {
		filter["genres"] = bson.M{"$in": []string{genre}}
	}
	if status != "" {
		filter["status"] = status
	}
	if author != "" {
		filter["author"] = bson.M{"$regex": author, "$options": "i"}
	}

	// Tạo options và gán trực tiếp sort ở đây (tránh SA4006)
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(limit64)

	switch sort {
	case "views_desc":
		findOptions.SetSort(bson.D{{Key: "view_count", Value: -1}})
	case "chapters_desc":
		findOptions.SetSort(bson.D{{Key: "chapters_count", Value: -1}})
	case "title_asc":
		findOptions.SetSort(bson.D{{Key: "title", Value: 1}})
	default:
		findOptions.SetSort(bson.D{{Key: "updated_at", Value: -1}})
	}

	// Count
	total, _ := storyCollection.CountDocuments(ctx, filter)

	// Query
	cursor, err := storyCollection.Find(ctx, filter, findOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể truy vấn"})
		return
	}
	defer cursor.Close(ctx)

	var stories []models.Story
	if err := cursor.All(ctx, &stories); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi đọc dữ liệu"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"page":    page,
		"limit":   limit,
		"total":   total,
		"stories": stories,
	})
}
func GetTopRankedStories(c *gin.Context) {
	storyCollection := config.MongoDB.Collection("Stories")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	limit64 := int64(limit)

	findOptions := options.Find().
		SetSort(bson.D{{Key: "view_count", Value: -1}}).
		SetLimit(limit64)

	cursor, err := storyCollection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		c.JSON(500, gin.H{"error": "Không thể truy vấn top truyện"})
		return
	}
	defer cursor.Close(ctx)

	var stories []models.Story
	if err := cursor.All(ctx, &stories); err != nil {
		c.JSON(500, gin.H{"error": "Lỗi đọc dữ liệu"})
		return
	}

	c.JSON(200, stories)
}

// GET /stories/:id/export
func ExportStoryChapters(c *gin.Context) {
	storyIDStr := c.Param("id")
	storyID, err := primitive.ObjectIDFromHex(storyIDStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "ID không hợp lệ"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	storyCollection := config.MongoDB.Collection("Stories")
	chapterCollection := config.MongoDB.Collection("Chapters")

	var story models.Story
	err = storyCollection.FindOne(ctx, bson.M{"_id": storyID}).Decode(&story)
	if err != nil {
		c.JSON(404, gin.H{"error": "Không tìm thấy truyện"})
		return
	}

	cursor, err := chapterCollection.Find(ctx, bson.M{"story_id": storyID})
	if err != nil {
		c.JSON(500, gin.H{"error": "Không thể truy vấn chương"})
		return
	}
	defer cursor.Close(ctx)

	var chapters []models.Chapter
	if err := cursor.All(ctx, &chapters); err != nil {
		c.JSON(500, gin.H{"error": "Lỗi đọc dữ liệu chương"})
		return
	}

	// Trả ra 1 gói JSON
	c.JSON(200, gin.H{
		"story":    story,
		"chapters": chapters,
	})
}
