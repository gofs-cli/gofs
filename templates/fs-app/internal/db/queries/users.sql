-- name: GetUsers :many
SELECT
    *
FROM
    users;

-- name: InsertUser :exec
INSERT INTO
    users (name, email)
VALUES
    (?, ?);
