-- name: GetLinks :many
SELECT
    id,
    original_url,
    short_name,
    short_url
FROM links;

-- name: CreateLink :one
INSERT INTO links(original_url, short_name, short_url)
VALUES ($1, $2, $3)
RETURNING id, original_url, short_name, short_url;

-- name: GetLinkByID :one
SELECT
    id,
    original_url,
    short_name,
    short_url
FROM links WHERE id = $1;

-- name: UpdateLinkByID :one
UPDATE links
SET original_url = $1, short_name = $2, short_url = $3
WHERE id = $4
RETURNING id, original_url, short_name, short_url;

-- name: DeleteLinkByID :execrows
DELETE FROM links WHERE id = $1;

