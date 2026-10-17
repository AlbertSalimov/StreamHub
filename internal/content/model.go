package content

import (
	"time"

	"github.com/google/uuid"
)

type Content struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Genre       string    `json:"genre"`
	ReleaseYear int       `json:"release_year"`
	IsPublished bool      `json:"is_published"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreateInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Genre       string `json:"genre"`
	ReleaseYear int    `json:"release_year"`
	IsPublished bool   `json:"is_published"`
}

type UpdateInput struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Genre       *string `json:"genre"`
	ReleaseYear *int    `json:"release_year"`
	IsPublished *bool   `json:"is_published"`
}
