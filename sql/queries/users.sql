-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email)
VALUES (
    gen_random_uuid (),
    $1,
    $2,
    $3
)
RETURNING *;

-- name: DestroyAllUsers :exec
DELETE FROM users;
