package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Chapter struct {
	Title   string `bson:"title" json:"title"`
	Content string `bson:"content" json:"content"`
}

type Story struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title       string             `bson:"title" json:"title"`
	Author      string             `bson:"author" json:"author"`
	Description string             `bson:"description" json:"description"`
	Genres      []string           `bson:"genres" json:"genres"`
	CoverURL    string             `bson:"cover_url" json:"cover_url"`
	Chapters    []Chapter          `bson:"chapters" json:"chapters"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}
