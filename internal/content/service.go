package content

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(
	ctx context.Context,
	input CreateInput,
) (Content, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	input.Genre = strings.TrimSpace(input.Genre)

	if err := validateFields(
		input.Title,
		input.Genre,
		input.ReleaseYear,
	); err != nil {
		return Content{}, err
	}

	return s.repo.Create(ctx, input)
}

func (s *Service) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (Content, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(
	ctx context.Context,
	limit, offset int,
) ([]Content, error) {
	if limit < 1 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.List(ctx, limit, offset)
}

func (s *Service) Update(
	ctx context.Context,
	id uuid.UUID,
	input UpdateInput,
) (Content, error) {
	if input.Title != nil {
		value := strings.TrimSpace(*input.Title)
		input.Title = &value
		if value == "" {
			return Content{}, fmt.Errorf("title must not be empty")
		}
	}

	if input.Genre != nil {
		value := strings.TrimSpace(*input.Genre)
		input.Genre = &value
		if value == "" {
			return Content{}, fmt.Errorf("genre must not be empty")
		}
	}

	if input.ReleaseYear != nil &&
		(*input.ReleaseYear < 1888 || *input.ReleaseYear > 2100) {
		return Content{}, fmt.Errorf("release_year must be between 1888 and 2100")
	}

	if input.Description != nil {
		value := strings.TrimSpace(*input.Description)
		input.Description = &value
	}

	return s.repo.Update(ctx, id, input)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func validateFields(title, genre string, year int) error {
	if title == "" {
		return fmt.Errorf("title is required")
	}
	if len(title) > 200 {
		return fmt.Errorf("title must be at most 200 characters")
	}
	if genre == "" {
		return fmt.Errorf("genre is required")
	}
	if len(genre) > 100 {
		return fmt.Errorf("genre must be at most 100 characters")
	}
	if year < 1888 || year > 2100 {
		return fmt.Errorf("release_year must be between 1888 and 2100")
	}

	return nil
}
