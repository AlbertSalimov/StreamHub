package content

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("content not found")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(
	ctx context.Context,
	input CreateInput,
) (Content, error) {
	const query = `
		INSERT INTO content (
			id, title, description, genre,
			release_year, is_published
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, title, description, genre,
		          release_year, is_published, created_at
	`

	id := uuid.New()

	var item Content
	err := r.db.QueryRow(
		ctx,
		query,
		id,
		input.Title,
		input.Description,
		input.Genre,
		input.ReleaseYear,
		input.IsPublished,
	).Scan(
		&item.ID,
		&item.Title,
		&item.Description,
		&item.Genre,
		&item.ReleaseYear,
		&item.IsPublished,
		&item.CreatedAt,
	)
	if err != nil {
		return Content{}, fmt.Errorf("create content: %w", err)
	}

	return item, nil
}

func (r *Repository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (Content, error) {
	const query = `
		SELECT id, title, description, genre,
		       release_year, is_published, created_at
		FROM content
		WHERE id = $1
	`

	var item Content
	err := r.db.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.Title,
		&item.Description,
		&item.Genre,
		&item.ReleaseYear,
		&item.IsPublished,
		&item.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Content{}, ErrNotFound
	}
	if err != nil {
		return Content{}, fmt.Errorf("get content: %w", err)
	}

	return item, nil
}

func (r *Repository) List(
	ctx context.Context,
	filter ListFilter,
) ([]Content, error) {
	query := `
		SELECT id, title, description, genre,
		       release_year, is_published, created_at
		FROM content
		WHERE is_published = TRUE
	`
	args := make([]any, 0, 5)
	addArg := func(value any) string {
		args = append(args, value)
		return fmt.Sprintf("$%d", len(args))
	}

	if filter.Search != "" {
		param := addArg("%" + filter.Search + "%")
		query += " AND title ILIKE " + param
	}

	if filter.Genre != "" {
		param := addArg(filter.Genre)
		query += " AND genre ILIKE " + param
	}

	if filter.Year != 0 {
		param := addArg(filter.Year)
		query += " AND release_year = " + param
	}

	sortColumns := map[string]string{
		"title": "title",
		"year":  "release_year",
		"date":  "created_at",
	}

	sortColumn, ok := sortColumns[filter.Sort]
	if !ok {
		sortColumn = "created_at"
	}

	sortOrder := "DESC"
	if filter.Order == "asc" {
		sortOrder = "ASC"
	}

	query += " ORDER BY " + sortColumn + " " + sortOrder + ", id ASC"
	query += " LIMIT " + addArg(filter.Limit)
	query += " OFFSET " + addArg(filter.Offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list content: %w", err)
	}
	defer rows.Close()

	items := make([]Content, 0)

	for rows.Next() {
		var item Content

		if err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.Description,
			&item.Genre,
			&item.ReleaseYear,
			&item.IsPublished,
			&item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan content: %w", err)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate content: %w", err)
	}

	return items, nil
}

func (r *Repository) Update(
	ctx context.Context,
	id uuid.UUID,
	input UpdateInput,
) (Content, error) {
	const query = `
		UPDATE content
		SET
			title = COALESCE($2, title),
			description = COALESCE($3, description),
			genre = COALESCE($4, genre),
			release_year = COALESCE($5, release_year),
			is_published = COALESCE($6, is_published)
		WHERE id = $1
		RETURNING id, title, description, genre,
		          release_year, is_published, created_at
	`

	var item Content
	err := r.db.QueryRow(
		ctx,
		query,
		id,
		input.Title,
		input.Description,
		input.Genre,
		input.ReleaseYear,
		input.IsPublished,
	).Scan(
		&item.ID,
		&item.Title,
		&item.Description,
		&item.Genre,
		&item.ReleaseYear,
		&item.IsPublished,
		&item.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Content{}, ErrNotFound
	}
	if err != nil {
		return Content{}, fmt.Errorf("update content: %w", err)
	}

	return item, nil
}

func (r *Repository) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM content WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete content: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
