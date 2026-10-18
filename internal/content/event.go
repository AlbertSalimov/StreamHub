package content

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID         uuid.UUID `json:"id"`
	Type       string    `json:"type"`
	OccurredAt time.Time `json:"occurred_at"`
	ContentID  uuid.UUID `json:"content_id"`
	Content    *Content  `json:"content,omitempty"`
}

type EventPublisher interface {
	Publish(ctx context.Context, event Event) error
}
