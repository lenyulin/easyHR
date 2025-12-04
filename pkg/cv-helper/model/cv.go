package model

import "time"

type CV struct {
	ID        string    `bson:"_id,omitempty" json:"id"`
	FilePath  string    `bson:"file_path" json:"file_path"`
	Content   string    `bson:"content" json:"content"`
	ParsedAt  time.Time `bson:"parsed_at" json:"parsed_at"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}
