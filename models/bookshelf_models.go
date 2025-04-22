package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BookshelfItem struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	StoryID   primitive.ObjectID `bson:"story_id" json:"story_id"`
	ChapterID primitive.ObjectID `bson:"last_chapter_id,omitempty" json:"last_chapter_id,omitempty"`
	AddedAt   time.Time          `bson:"added_at" json:"added_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}
