-- name: UpsertAccountLimit :one
INSERT INTO account_limits (
  account_id,
  daily_transfer_limit
) VALUES (
  $1, $2
)
ON CONFLICT (account_id)
DO UPDATE SET
  daily_transfer_limit = EXCLUDED.daily_transfer_limit,
  updated_at = now()
RETURNING *;

-- name: GetAccountLimit :one
SELECT * FROM account_limits
WHERE account_id = $1 LIMIT 1;
