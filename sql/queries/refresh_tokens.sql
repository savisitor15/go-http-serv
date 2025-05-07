-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at)
VALUES (
    $1,
    $2,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: GetToken :one
SELECT * FROM refresh_tokens
WHERE token = $1;

-- name: ExpireToken :exec
UPDATE refresh_tokens
SET 
revoked_at = $2,
updated_at = $2
WHERE token = $1;

-- name: RemoveToken :exec
DELETE FROM refresh_tokens
WHERE token = $1;

-- name: DestroyAllTokens :exec
DELETE FROM refresh_tokens;