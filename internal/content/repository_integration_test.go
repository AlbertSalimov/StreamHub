package content

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestRepositoryIntegration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set; skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to test database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping test database: %v", err)
	}

	_, err = pool.Exec(ctx, `DROP TABLE IF EXISTS content`)
	if err != nil {
		t.Fatalf("drop test table: %v", err)
	}

	_, err = pool.Exec(ctx, `
	CREATE TABLE content (
		id UUID PRIMARY KEY,
		title TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		genre TEXT NOT NULL,
		release_year INT NOT NULL
			CHECK (release_year BETWEEN 1888 AND 2100),
		is_published BOOLEAN NOT NULL DEFAULT false,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now()
	)
`)
	if err != nil {
		t.Fatalf("create test table: %v", err)
	}

	repo := NewRepository(pool)

	t.Run("create and get", func(t *testing.T) {
		input := CreateInput{
			Title:       "Integration Test Movie",
			Description: "Created by Go test",
			Genre:       "drama",
			ReleaseYear: 2024,
			IsPublished: true,
		}

		created, err := repo.Create(ctx, input)
		if err != nil {
			t.Fatalf("create content: %v", err)
		}

		if created.ID == uuid.Nil {
			t.Fatal("expected generated UUID")
		}
		if created.Title != input.Title {
			t.Fatalf("title = %q, want %q", created.Title, input.Title)
		}

		got, err := repo.GetByID(ctx, created.ID)
		if err != nil {
			t.Fatalf("get content: %v", err)
		}
		if got.ID != created.ID {
			t.Fatalf("id = %s, want %s", got.ID, created.ID)
		}
	})

	t.Run("list only published content", func(t *testing.T) {
		_, err := repo.Create(ctx, CreateInput{
			Title:       "Unpublished Movie",
			Genre:       "drama",
			ReleaseYear: 2023,
			IsPublished: false,
		})
		if err != nil {
			t.Fatalf("create unpublished content: %v", err)
		}

		items, err := repo.List(ctx, ListFilter{
			Limit:  20,
			Offset: 0,
			Sort:   "title",
			Order:  "asc",
		})
		if err != nil {
			t.Fatalf("list content: %v", err)
		}

		for _, item := range items {
			if !item.IsPublished {
				t.Errorf("unpublished content %s appeared in public list", item.ID)
			}
		}
	})

	t.Run("delete missing content", func(t *testing.T) {
		err := repo.Delete(ctx, uuid.New())
		if err != ErrNotFound {
			t.Fatalf("delete error = %v, want %v", err, ErrNotFound)
		}
	})
}
