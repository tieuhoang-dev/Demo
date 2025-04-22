package middlewares

import (
	"Truyen_BE/config"
	"Truyen_BE/models"
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var jwtSecret = []byte("your-secret-key") // nên dùng từ ENV

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Lấy token từ header
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Thiếu hoặc sai định dạng token"})
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 2. Parse token
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token không hợp lệ"})
			return
		}

		// 3. Lấy claims từ token
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token lỗi claims"})
			return
		}

		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token thiếu user_id"})
			return
		}
		userID, err := primitive.ObjectIDFromHex(userIDStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token user_id không hợp lệ"})
			return
		}

		// 4. Truy user từ DB
		userCollection := config.MongoDB.Collection("Users")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var user models.User
		err = userCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy người dùng"})
			return
		}

		// 5. Kiểm tra status
		if user.Status == "banned" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Tài khoản đã bị khóa"})
			return
		}

		// 6. Lưu thông tin user vào context
		c.Set("user_id", user.ID)
		c.Set("user_role", user.Role)
		c.Set("username", user.Username)

		// 7. Cho đi tiếp
		c.Next()
	}
}
