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

// Middleware: Kiểm tra user có phải tác giả của chương (suy ra từ truyện)
func IsAuthorOfChapter() gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Chưa đăng nhập"})
			c.Abort()
			return
		}
		userID := userIDVal.(primitive.ObjectID)

		chapterIDStr := c.Param("id")
		chapterID, err := primitive.ObjectIDFromHex(chapterIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID chương không hợp lệ"})
			c.Abort()
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		chapterCollection := config.MongoDB.Collection("Chapters")
		storyCollection := config.MongoDB.Collection("Stories")

		var chapter struct {
			StoryID primitive.ObjectID `bson:"story_id"`
		}
		err = chapterCollection.FindOne(ctx, bson.M{"_id": chapterID}).Decode(&chapter)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy chương"})
			c.Abort()
			return
		}

		var story struct {
			CreatedBy primitive.ObjectID `bson:"created_by"`
		}
		err = storyCollection.FindOne(ctx, bson.M{"_id": chapter.StoryID}).Decode(&story)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy truyện gốc của chương"})
			c.Abort()
			return
		}

		if story.CreatedBy != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không phải tác giả của chương này"})
			c.Abort()
			return
		}

		c.Next()
	}
}
