package controllers

import (
	"Truyen_BE/config"
	"Truyen_BE/models"
	"context"
	"fmt"
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
	filter := bson.M{
		"is_hidden": false,
		"is_banned": false,
	}
	cursor, err := storyCollection.Find(ctx, filter)
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
	newStory.IsBanned = false
	newStory.DeletedAt = nil
	newStory.CreatedBy = c.MustGet("user_id").(primitive.ObjectID)
	newStory.Status = "active" // Hoặc trạng thái mặc định khác
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	val, exists := c.Get("user_id")
	if !exists {
		fmt.Println("❌ Controller không thấy user_id trong context")
	} else {
		fmt.Println("✅ Controller lấy user_id:", val)
	}
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
	bookshelfCollection := config.MongoDB.Collection("Bookshelf")
	// Xoá truyện trong tủ sách
	_, err = bookshelfCollection.DeleteMany(ctx, bson.M{"story_id": objectID})
	if err != nil {
		c.JSON(500, gin.H{"error": "Không thể xoá truyện trong tủ sách"})
		return
	}

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

// GET /stories/featured
func GetFeaturedStories(c *gin.Context) {
	storyCollection := config.MongoDB.Collection("Stories")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := storyCollection.Find(ctx, bson.M{"is_featured": true})
	if err != nil {
		c.JSON(500, gin.H{"error": "Không thể truy vấn truyện đề cử"})
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
func BanStory(c *gin.Context) {
	title := c.Param("title")
	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu tên truyện"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	storyCollection := config.MongoDB.Collection("Stories")

	// Gán trạng thái bị ban
	update := bson.M{
		"$set": bson.M{
			"is_banned":  true,
			"deleted_at": time.Now(),
		},
	}

	result, err := storyCollection.UpdateOne(ctx, bson.M{"title": title}, update)
	if err != nil || result.MatchedCount == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể ban hoặc không tìm thấy truyện"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "✅ Truyện đã bị ẩn (soft-delete)"})
}
func UnbanStory(c *gin.Context) {
	title := c.Param("title")
	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu tên truyện"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	storyCollection := config.MongoDB.Collection("Stories")

	// Gán trạng thái không bị ban
	update := bson.M{
		"$set": bson.M{
			"is_banned":  false,
			"deleted_at": nil,
		},
	}

	result, err := storyCollection.UpdateOne(ctx, bson.M{"title": title}, update)
	if err != nil || result.MatchedCount == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể bỏ ban hoặc không tìm thấy truyện"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "✅ Truyện đã được bỏ ban"})
}
func DeleteStoryByAuthor(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không xác định được người dùng"})
		return
	}
	userID := userIDValue.(primitive.ObjectID)

	title := c.Param("title")
	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu tên truyện"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	storyCollection := config.MongoDB.Collection("Stories")
	chapterCollection := config.MongoDB.Collection("Chapters")
	bookshelfCollection := config.MongoDB.Collection("Bookshelf")

	// 1. Kiểm tra truyện có tồn tại và đúng tác giả không
	var story models.Story
	err := storyCollection.FindOne(ctx, bson.M{
		"title":      title,
		"created_by": userID,
		"is_banned":  true, // ❗ Chỉ cho xóa nếu đã bị ẩn trước
	}).Decode(&story)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không có quyền xóa truyện này hoặc truyện chưa bị ẩn"})
		return
	}

	// 2. Xóa chương liên quan
	_, _ = chapterCollection.DeleteMany(ctx, bson.M{"story_id": story.ID})

	// 3. Xóa khỏi tủ sách người dùng
	_, _ = bookshelfCollection.DeleteMany(ctx, bson.M{"story_id": story.ID})

	// 4. Xóa truyện
	_, err = storyCollection.DeleteOne(ctx, bson.M{"_id": story.ID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể xóa truyện"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "✅ Truyện đã được xóa vĩnh viễn"})
}
func GetStoryContent(c *gin.Context) {
	storyIDStr := c.Param("id")
	storyID, err := primitive.ObjectIDFromHex(storyIDStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "ID không hợp lệ"})
		return
	}

	// Log ID truy vấn để kiểm tra
	fmt.Println("Truy vấn với ID:", storyIDStr)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Kiểm tra lại filter
	filter := bson.M{
		"_id": storyID,
	}

	storyCollection := config.MongoDB.Collection("Stories")
	var story models.Story
	err = storyCollection.FindOne(ctx, filter).Decode(&story)
	if err != nil {
		fmt.Println("Lỗi truy vấn:", err)
		c.JSON(404, gin.H{"error": "Không tìm thấy truyện"})
		return
	}

	c.JSON(200, story)
}
func GetGenresWithCount(c *gin.Context) {
	storyCollection := config.MongoDB.Collection("Stories")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pipeline := []bson.M{
		{"$unwind": "$genres"},
		{"$group": bson.M{
			"_id":   "$genres",
			"count": bson.M{"$sum": 1},
		}},
		{"$sort": bson.M{"count": -1}},
	}

	cursor, err := storyCollection.Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể truy vấn thể loại"})
		return
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi đọc kết quả"})
		return
	}

	c.JSON(http.StatusOK, results)
}
func GetAllStoriesOfGenre(c *gin.Context) {
	// Lấy thể loại từ URL
	genre := c.Param("genre")
	if genre == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu thể loại"})
		return
	}

	// Truy vấn đến MongoDB
	storyCollection := config.MongoDB.Collection("Stories")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Tạo bộ lọc để tìm truyện theo thể loại
	filter := bson.M{"genres": genre}
	cursor, err := storyCollection.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể truy vấn truyện"})
		return
	}
	defer cursor.Close(ctx)

	var stories []models.Story
	if err := cursor.All(ctx, &stories); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi đọc dữ liệu"})
		return
	}

	// Trả về kết quả truy vấn
	c.JSON(http.StatusOK, stories)
}
func GetNewestUpdatedStoryList(c *gin.Context) {
	storyCollection := config.MongoDB.Collection("Stories")
	chapterCollection := config.MongoDB.Collection("Chapters")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	limit64 := int64(limit)

	findOptions := options.Find().
		SetSort(bson.D{{Key: "updated_at", Value: -1}}).
		SetLimit(limit64)

	cursor, err := storyCollection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể truy vấn truyện mới nhất"})
		return
	}
	defer cursor.Close(ctx)

	var stories []models.Story
	if err := cursor.All(ctx, &stories); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi đọc dữ liệu"})
		return
	}

	var result []models.StoryWithLatestChapter

	for _, story := range stories {
		var latestChapter models.Chapter
		err := chapterCollection.FindOne(ctx,
			bson.M{"story_id": story.ID},
			options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}}),
		).Decode(&latestChapter)

		storyWithChapter := models.StoryWithLatestChapter{
			Story:         story,
			LatestChapter: nil,
		}

		if err == nil {
			storyWithChapter.LatestChapter = &latestChapter
		}

		result = append(result, storyWithChapter)
	}

	c.JSON(http.StatusOK, result)
}
