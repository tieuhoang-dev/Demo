package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client
var MongoDB *mongo.Database

func LoadEnv() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal("❌ Không thể lấy thư mục hiện tại:", err)
	}
	fmt.Println("📂 Đang chạy tại thư mục:", dir)

	err = godotenv.Load()
	if err != nil {
		log.Fatal("❌ Không thể load file .env:", err)
	}
	fmt.Println("✅ Đã load file .env")
}

func ConnectDB() {
	uri := os.Getenv("MONGO_URI")
	dbName := os.Getenv("DB_NAME")

	// Tạo context KHÔNG có timeout, vì client cần sống suốt runtime
	ctx := context.Background()

	// Kết nối MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal("❌ Lỗi kết nối MongoDB:", err)
	}

	// Kiểm tra kết nối
	ctxPing, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Ping(ctxPing, nil)
	if err != nil {
		log.Fatal("❌ Không thể ping MongoDB:", err)
	}

	MongoClient = client
	MongoDB = client.Database(dbName)
	log.Println("✅ Đã kết nối MongoDB thành công!")
}
