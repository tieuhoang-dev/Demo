package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Chapter struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	StoryID       primitive.ObjectID `bson:"story_id" json:"story_id"` // Liên kết với Story
	ChapterNumber int                `bson:"chapter_number" json:"chapter_number"`
	Title         string             `bson:"title" json:"title"`
	Content       string             `bson:"content" json:"content"`
	ViewCount     int64              `bson:"view_count" json:"view_count"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}

type Story struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title         string             `bson:"title" json:"title"`
	Author        string             `bson:"author" json:"author"`
	Description   string             `bson:"description" json:"description"`
	CoverURL      string             `bson:"cover_url" json:"cover_url"`
	Genres        []string           `bson:"genres" json:"genres"` // Ví dụ: ["Tiên hiệp", "Hài hước"]
	Status        string             `bson:"status" json:"status"` // "ongoing", "completed"
	ChaptersCount int                `bson:"chapters_count" json:"chapters_count"`
	ViewCount     int64              `bson:"view_count" json:"view_count"`
	IsFeatured    bool               `bson:"is_featured" json:"is_featured"`
	IsHidden      bool               `bson:"is_hidden" json:"is_hidden"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
	IsBanned      bool               `bson:"is_banned" json:"is_banned"`
	DeletedAt     *time.Time         `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
	CreatedBy     primitive.ObjectID `bson:"created_by" json:"created_by"`
}
type StoryWithLatestChapter struct {
	Story
	LatestChapter *Chapter `json:"latest_chapter,omitempty"`
}
