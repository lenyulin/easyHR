package cv

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CV represents a curriculum vitae document and its metadata.
type CV struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	FilePath  string             `bson:"file_path" json:"file_path"`
	Content   string             `bson:"content" json:"content"`
	ParsedAt  time.Time          `bson:"parsed_at" json:"parsed_at"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// Storage defines the interface for persisting data.
type Storage interface {
	Insert(ctx context.Context, collection string, doc interface{}) error
}

// Parser defines the interface for parsing CV documents.
type Parser interface {
	Parse(filePath string) (string, error)
}
