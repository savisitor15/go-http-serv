-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid (),
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: UpdateUserByID :one
UPDATE users
SET
email = $2,
hashed_password = $3
WHERE id = $1
RETURNING *;

-- name: SetChirpyRedByID :one
UPDATE users
SET
is_chirpy_red = true
WHERE id = $1
RETURNING *;

-- name: UnSetChirpyRedByID :one
UPDATE users
SET
is_chirpy_red = false
WHERE id = $1
RETURNING *;

-- name: DestroyAllUsers :exec
DELETE FROM users;
