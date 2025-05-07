package controllers

import (
	"Truyen_BE/config"
	"Truyen_BE/models"
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"

)

func RegisterUser(c *gin.Context) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email,omitempty"`
	}

	// Bước 1: Validate đầu vào
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu đầu vào không hợp lệ"})
		return
	}

	input.Username = strings.TrimSpace(input.Username)
	input.Email = strings.TrimSpace(input.Email)

	if input.Username == "" || input.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username và Password là bắt buộc"})
		return
	}

	// Bước 2: Kiểm tra username đã tồn tại chưa
	userCollection := config.MongoDB.Collection("Users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	count, err := userCollection.CountDocuments(ctx, bson.M{"username": input.Username})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi kiểm tra username"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Username đã tồn tại"})
		return
	}

	// Bước 3: Hash mật khẩu
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể mã hóa mật khẩu"})
		return
	}

	// Bước 4: Tạo User mới
	newUser := models.User{
		ID:        primitive.NewObjectID(),
		Username:  input.Username,
		Email:     input.Email,
		Password:  string(hashedPassword),
		Role:      "user",   // mặc định
		Status:    "active", // chưa bị ban
		CreatedAt: time.Now(),
	}

	// Bước 5: Lưu vào MongoDB
	_, err = userCollection.InsertOne(ctx, newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tạo tài khoản"})
		return
	}

	// Trả kết quả (ẩn pass)
	c.JSON(http.StatusOK, gin.H{
		"message": "✅ Đăng ký thành công",
		"user": gin.H{
			"id":       newUser.ID.Hex(),
			"username": newUser.Username,
			"role":     newUser.Role,
			"status":   newUser.Status,
		},
	})
}
func GetCurrentUser(c *gin.Context) {
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")
	role, _ := c.Get("user_role")
	created_at, _ := c.Get("created_at")
	status, _ := c.Get("status")

	c.JSON(http.StatusOK, gin.H{
		"user_id":  userID,
		"username": username,
		"role":     role,
		"created_at": created_at,
		"status":   status,
	})
}

// API handler để lấy truyện của người dùng
func GetUserStories(c *gin.Context) {
    userIDVal, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Chưa đăng nhập"})
        return
    }

    userID, ok := userIDVal.(primitive.ObjectID)
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "ID người dùng không hợp lệ"})
        return
    }

    // Truy vấn MongoDB để lấy danh sách truyện của người dùng
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    var stories []models.Story
    cursor, err := config.MongoDB.Collection("Stories").Find(ctx, bson.M{"created_by": userID})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy truyện"})
        return
    }
    if err := cursor.All(ctx, &stories); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy dữ liệu truyện"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"stories": stories})
}
