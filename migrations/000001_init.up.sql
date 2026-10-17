CREATE TABLE content (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    genre TEXT NOT NULL,
    release_year INT NOT NULL
        CHECK (release_year BETWEEN 1888 AND 2100),
    is_published BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_content_published_year
    ON content (release_year DESC, id)
    WHERE is_published = true;