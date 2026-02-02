-- name: CreateVisit :exec
INSERT INTO visits (link_id, ip, user_agent, referer, status)
VALUES ($1, $2, $3, $4, $5);

-- name: GetVisits :many
SELECT
    id,
    link_id,
    created_at,
    ip,
    user_agent,
    status
FROM visits
ORDER BY created_at DESC, id DESC
LIMIT $1 OFFSET $2;

-- name: GetTotalVisits :one
SELECT COUNT(id) AS total_visits FROM visits;