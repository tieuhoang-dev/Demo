package controllers

import (
	"Truyen_BE/config"
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func RemoveFromBookshelf(c *gin.Context) {
	// 1. Lấy user_id từ context (middleware đã gán)
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy user trong context"})
		return
	}
	userID := userIDValue.(primitive.ObjectID)

	// 2. Nhận story_id từ URL
	storyID := c.Param("story_id")
	if storyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu story_id trong URL"})
		return
	}

	// 3. Chuyển story_id sang ObjectID
	storyObjectID, err := primitive.ObjectIDFromHex(storyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "story_id không hợp lệ"})
		return
	}

	// 4. Xóa khỏi tủ sách
	collection := config.MongoDB.Collection("Bookshelf")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.DeleteOne(ctx, bson.M{
		"user_id":  userID,
		"story_id": storyObjectID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể xóa khỏi tủ sách"})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy truyện trong tủ sách"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "✅ Đã xóa truyện khỏi tủ sách"})
}

func GetBookshelf(c *gin.Context) {
	// Lấy thông tin userID từ context
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy user trong context"})
		return
	}
	userID := userIDValue.(primitive.ObjectID)

	// Lấy các tham số phân trang và sắp xếp từ query params
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")
	sortBy := c.DefaultQuery("sortBy", "updated_at")
	sortOrder := c.DefaultQuery("sortOrder", "-1")

	// Chuyển đổi các giá trị query thành kiểu phù hợp
	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Trang không hợp lệ"})
		return
	}
	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Số lượng kết quả không hợp lệ"})
		return
	}
	sortOrderInt, err := strconv.Atoi(sortOrder)
	if err != nil || (sortOrderInt != 1 && sortOrderInt != -1) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thứ tự sắp xếp không hợp lệ"})
		return
	}

	collection := config.MongoDB.Collection("Bookshelf")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Xây dựng pipeline cho aggregation
	pipeline := mongo.Pipeline{
		// Lọc theo user_id
		{{Key: "$match", Value: bson.D{{Key: "user_id", Value: userID}}}},

		// Join với collection "stories"
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Stories"},
			{Key: "localField", Value: "story_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "story"},
		}}},

		{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$story"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}},

		// Join với collection "chapters"
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Chapters"},
			{Key: "localField", Value: "last_chapter_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "chapter"},
		}}},

		{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$chapter"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}},

		// Kiểm tra và thay thế null hoặc giá trị không hợp lệ cho trường "updated_at"
		{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "story_id", Value: "$story._id"},
			{Key: "story_title", Value: "$story.title"},
			{Key: "last_chapter_id", Value: "$chapter._id"},
			{Key: "chapter_number", Value: "$chapter.chapter_number"},
			{Key: "chapter_title", Value: "$chapter.title"},
			{Key: "added_at", Value: 1},
			{Key: "updated_at", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$updated_at", "2025-04-22"}}}}, // Giá trị mặc định nếu null
		}}},

		// Thêm sắp xếp theo trường "sortBy" và "sortOrder"
		{{Key: "$sort", Value: bson.D{{Key: sortBy, Value: sortOrderInt}}}},
	}

	// Thêm phân trang vào pipeline
	skip := (pageInt - 1) * limitInt
	pipeline = append(pipeline, bson.D{{Key: "$skip", Value: skip}}, bson.D{{Key: "$limit", Value: limitInt}})

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy tủ sách"})
		return
	}
	defer cursor.Close(ctx)

	var result []bson.M
	if err = cursor.All(ctx, &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi xử lý kết quả"})
		return
	}

	// Nếu không có kết quả, trả về mảng rỗng
	if len(result) == 0 {
		c.JSON(http.StatusOK, []bson.M{})
		return
	}

	c.JSON(http.StatusOK, result)
}

func UpdateLastChapter(c *gin.Context) {
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
		LastChapterID string `json:"last_chapter_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil || input.StoryID == "" || input.LastChapterID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu hoặc sai định dạng story_id hoặc last_chapter_id"})
		return
	}

	// 3. Chuyển story_id và last_chapter_id sang ObjectID
	storyObjectID, err := primitive.ObjectIDFromHex(input.StoryID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "story_id không hợp lệ"})
		return
	}
	lastChapterObjectID, err := primitive.ObjectIDFromHex(input.LastChapterID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "last_chapter_id không hợp lệ"})
		return
	}

	// 4. Kiểm tra xem câu chuyện đã có trong tủ sách chưa
	collection := config.MongoDB.Collection("Bookshelf")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var bookshelfItem bson.M
	err = collection.FindOne(ctx, bson.M{
		"user_id":  userID,
		"story_id": storyObjectID,
	}).Decode(&bookshelfItem)

	// Nếu không có, thêm mới vào tủ sách
	if err == mongo.ErrNoDocuments {
		_, err = collection.InsertOne(ctx, bson.M{
			"user_id":         userID,
			"story_id":        storyObjectID,
			"chapter_id":      lastChapterObjectID,
			"added_at":        time.Now(),
			"updated_at":      time.Now(),
			"last_chapter_id": lastChapterObjectID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể thêm câu chuyện vào tủ sách"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "✅ Đã thêm câu chuyện vào tủ sách"})
		return
	}

	// Nếu có, cập nhật lại last_chapter_id
	_, err = collection.UpdateOne(ctx, bson.M{
		"user_id":  userID,
		"story_id": storyObjectID,
	}, bson.M{
		"$set": bson.M{
			"chapter_id":      lastChapterObjectID,
			"last_chapter_id": lastChapterObjectID, // thêm dòng này
			"updated_at":      time.Now(),
		},
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể cập nhật chương cuối"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "✅ Đã cập nhật chương cuối"})
}
