package content

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Cache interface {
	Get(ctx context.Context, key string) ([]byte, bool, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	DeleteByPrefix(ctx context.Context, prefix string) error
}

type Service struct {
	repo  *Repository
	cache Cache
	ttl   time.Duration
}

func NewService(repo *Repository, caches ...Cache) *Service {
	service := &Service{
		repo: repo,
		ttl:  60 * time.Second,
	}

	if len(caches) > 0 {
		service.cache = caches[0]
	}

	return service
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

	item, err := s.repo.Create(ctx, input)
	if err != nil {
		return Content{}, err
	}

	s.invalidateListCache(ctx)
	return item, nil
}

func (s *Service) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (Content, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(
	ctx context.Context,
	filter ListFilter,
) ([]Content, error) {
	filter.Search = strings.TrimSpace(filter.Search)
	filter.Genre = strings.TrimSpace(filter.Genre)

	if len(filter.Search) > 100 {
		return nil, fmt.Errorf("search must be at most 100 characters")
	}
	if len(filter.Genre) > 100 {
		return nil, fmt.Errorf("genre must be at most 100 characters")
	}

	if filter.Year != 0 && (filter.Year < 1888 || filter.Year > 2100) {
		return nil, fmt.Errorf("year must be between 1888 and 2100")
	}

	if filter.Limit == 0 {
		filter.Limit = 20
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		return nil, fmt.Errorf("limit must be between 1 and 100")
	}
	if filter.Offset < 0 {
		return nil, fmt.Errorf("offset must be zero or greater")
	}

	if filter.Sort == "" {
		filter.Sort = "date"
	}

	switch filter.Sort {
	case "title", "year", "date":
	default:
		return nil, fmt.Errorf("sort must be title, year, or date")
	}

	if filter.Order == "" {
		filter.Order = "desc"
	}
	if filter.Order != "asc" && filter.Order != "desc" {
		return nil, fmt.Errorf("order must be asc or desc")
	}

	key := s.cacheKey(filter)

	if s.cache != nil {
		cacheCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		data, found, err := s.cache.Get(cacheCtx, key)
		cancel()

		if err == nil && found {
			var items []Content
			if json.Unmarshal(data, &items) == nil {
				return items, nil
			}
		}
	}

	items, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		if data, marshalErr := json.Marshal(items); marshalErr == nil {
			cacheCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			_ = s.cache.Set(cacheCtx, key, data, s.ttl)
			cancel()
		}
	}

	return items, nil
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

	item, err := s.repo.Update(ctx, id, input)
	if err != nil {
		return Content{}, err
	}

	s.invalidateListCache(ctx)
	return item, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	s.invalidateListCache(ctx)
	return nil
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

func (s *Service) cacheKey(filter ListFilter) string {
	data, _ := json.Marshal(filter)
	return "content:list:" + string(data)
}

func (s *Service) invalidateListCache(_ context.Context) {
	if s.cache == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_ = s.cache.DeleteByPrefix(ctx, "content:list:")
}
