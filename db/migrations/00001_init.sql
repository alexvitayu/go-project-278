-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS links(
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    original_url TEXT UNIQUE NOT NULL,
    short_name VARCHAR(100) UNIQUE NOT NULL,
    short_url TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS links;
-- +goose StatementEnd
