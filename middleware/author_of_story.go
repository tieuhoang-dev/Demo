package middlewares

import (
	"Truyen_BE/config"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Middleware: Kiểm tra user có phải tác giả của story (dựa vào story_id trong body)
func IsAuthorOfStory() gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Chưa đăng nhập"})
			c.Abort()
			return
		}
		userID := userIDVal.(primitive.ObjectID)

		var body struct {
			StoryID string `json:"story_id"`
		}
		if err := c.ShouldBindJSON(&body); err != nil || body.StoryID == "" {
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
