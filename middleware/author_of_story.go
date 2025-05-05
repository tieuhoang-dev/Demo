package middlewares

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"Truyen_BE/config"
)

func IsAuthorOfStory() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Lấy user_id từ context (đã được middleware xác thực trước đó)
		userIDVal, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Chưa đăng nhập"})
			c.Abort()
			return
		}
		userID, ok := userIDVal.(primitive.ObjectID)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "ID người dùng không hợp lệ"})
			c.Abort()
			return
		}

		// Đọc body một lần duy nhất
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Không đọc được dữ liệu"})
			c.Abort()
			return
		}

		// Khôi phục lại body cho các middleware hoặc handler tiếp theo
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Parse để kiểm tra story_id
		var body struct {
			StoryID string `json:"story_id"`
		}
		if err := json.Unmarshal(bodyBytes, &body); err != nil || body.StoryID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu hoặc sai định dạng story_id"})
			c.Abort()
			return
		}

		storyID, err := primitive.ObjectIDFromHex(body.StoryID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "story_id không hợp lệ"})
			c.Abort()
			return
		}

		// Truy vấn MongoDB để kiểm tra người tạo
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var result struct {
			CreatedBy primitive.ObjectID `bson:"created_by"`
		}
		err = config.MongoDB.Collection("Stories").FindOne(ctx, bson.M{"_id": storyID}).Decode(&result)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy truyện"})
			c.Abort()
			return
		}

		if result.CreatedBy != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không phải tác giả truyện này"})
			c.Abort()
			return
		}

		c.Next()
	}
}
