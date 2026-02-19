-- name: CreateEntry :one
INSERT INTO entries (
  account_id,
  amount
) VALUES (
  $1, $2
) RETURNING *;

-- name: GetEntry :one
SELECT * FROM entries
WHERE id = $1 LIMIT 1;

-- name: ListEntries :many
SELECT * FROM entries
WHERE account_id = $1
ORDER BY id
LIMIT $2
OFFSET $3;

-- name: ListEntriesFilteredDesc :many
SELECT * FROM entries
WHERE account_id = $1
  AND amount >= $2
  AND amount <= $3
  AND created_at >= $4
  AND created_at <= $5
ORDER BY created_at DESC
LIMIT $6
OFFSET $7;

-- name: ListEntriesFilteredAsc :many
SELECT * FROM entries
WHERE account_id = $1
  AND amount >= $2
  AND amount <= $3
  AND created_at >= $4
  AND created_at <= $5
ORDER BY created_at ASC
LIMIT $6
OFFSET $7;
