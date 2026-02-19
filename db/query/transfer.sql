-- name: CreateTransfer :one
INSERT INTO transfers (
  from_account_id,
  to_account_id,
  amount
) VALUES (
  $1, $2, $3
) RETURNING *;

-- name: GetTransfer :one
SELECT * FROM transfers
WHERE id = $1 LIMIT 1;

-- name: ListTransfers :many
SELECT * FROM transfers
WHERE 
    from_account_id = $1 OR
    to_account_id = $2
ORDER BY id
LIMIT $3
OFFSET $4;

-- name: ListTransfersFilteredDesc :many
SELECT * FROM transfers
WHERE (
    ($2 = 'any' AND (from_account_id = $1 OR to_account_id = $1)) OR
    ($2 = 'out' AND from_account_id = $1) OR
    ($2 = 'in' AND to_account_id = $1)
  )
  AND amount >= $3
  AND amount <= $4
  AND created_at >= $5
  AND created_at <= $6
ORDER BY created_at DESC
LIMIT $7
OFFSET $8;

-- name: ListTransfersFilteredAsc :many
SELECT * FROM transfers
WHERE (
    ($2 = 'any' AND (from_account_id = $1 OR to_account_id = $1)) OR
    ($2 = 'out' AND from_account_id = $1) OR
    ($2 = 'in' AND to_account_id = $1)
  )
  AND amount >= $3
  AND amount <= $4
  AND created_at >= $5
  AND created_at <= $6
ORDER BY created_at ASC
LIMIT $7
OFFSET $8;

-- name: GetDailyTransferTotal :one
SELECT COALESCE(SUM(amount), 0)::bigint AS total
FROM transfers
WHERE from_account_id = $1
  AND created_at >= $2
  AND created_at < $3;
