package models
import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive")
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username  string             `bson:"username" json:"username"`
	Email     string             `bson:"email,omitempty" json:"email,omitempty"`
	Password  string             `bson:"password,omitempty" json:"-"` // không trả password ra 
	Status string `bson:"status" json:"status"` // "active", "inactive", "banned"
	Role      string             `bson:"role" json:"role"`             // "user", "author", "admin"
	AvatarURL string             `bson:"avatar_url,omitempty" json:"avatar_url,omitempty"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}