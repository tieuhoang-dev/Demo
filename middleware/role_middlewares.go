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

// Middleware: Chỉ cho phép tác giả của truyện và admin thao tác
func AuthorOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Chưa xác thực người dùng"})
			c.Abort()
			return
		}
		userID := userIDVal.(primitive.ObjectID)

		role, _ := c.Get("user_role")
		if role == "admin" {
			c.Next() // ✅ Admin luôn có quyền
			return
		}

		// Lấy ID hoặc title tùy route
		storyIDParam := c.Param("id")
		storyTitle := c.Param("title")

		filter := bson.M{}
		if storyIDParam != "" {
			if objectID, err := primitive.ObjectIDFromHex(storyIDParam); err == nil {
				filter["_id"] = objectID
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "ID truyện không hợp lệ"})
				c.Abort()
				return
			}
		} else if storyTitle != "" {
			filter["title"] = storyTitle
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu thông tin truyện"})
			c.Abort()
			return
		}

		// Truy vấn DB
		var story struct {
			CreatedBy primitive.ObjectID `bson:"created_by"`
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		storyCollection := config.MongoDB.Collection("Stories")
		err := storyCollection.FindOne(ctx, filter).Decode(&story)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy truyện"})
			c.Abort()
			return
		}

		if story.CreatedBy != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không phải tác giả truyện này"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists || role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không có quyền admin"})
			c.Abort()
			return
		}
		c.Next()
	}
}
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Chưa xác thực người dùng"})
			c.Abort()
			return
		}

		roleValid := false
		for _, role := range allowedRoles {
			if userRole == role {
				roleValid = true
				break
			}
		}

		if !roleValid {
			c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không có quyền truy cập"})
			c.Abort()
			return
		}
		c.Next()
	}
}
