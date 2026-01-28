-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS visits (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    link_id BIGINT NOT NULL,
    ip VARCHAR(255) NOT NULL,
    user_agent TEXT NOT NULL,
    referer TEXT,
    status INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS visits;
-- +goose StatementEnd
