package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Chapter struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	StoryID        primitive.ObjectID `bson:"storyId"`
	Chapter_Number int                `bson:"chapterNumber"`
	Title          string             `bson:"title"`
	Content        string             `bson:"content"`
	CreatedAt      time.Time          `bson:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at"`
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
