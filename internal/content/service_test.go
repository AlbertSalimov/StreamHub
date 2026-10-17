package content

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestCreateRejectsMissingTitle(t *testing.T) {
	service := NewService(nil)

	_, err := service.Create(context.Background(), CreateInput{
		Genre:       "drama",
		ReleaseYear: 2024,
	})
	if err == nil {
		t.Fatal("expected an error when title is missing")
	}
}

func TestCreateRejectsInvalidYear(t *testing.T) {
	service := NewService(nil)

	_, err := service.Create(context.Background(), CreateInput{
		Title:       "Example",
		Genre:       "drama",
		ReleaseYear: 1700,
	})
	if err == nil {
		t.Fatal("expected an error for an invalid release year")
	}
}

func TestUpdateRejectsEmptyTitle(t *testing.T) {
	service := NewService(nil)
	empty := "   "

	_, err := service.Update(
		context.Background(),
		uuid.Nil,
		UpdateInput{Title: &empty},
	)
	if err == nil {
		t.Fatal("expected an error when title is empty")
	}
}
