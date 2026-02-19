-- name: UpsertAccountAlert :one
INSERT INTO account_alerts (
  account_id,
  low_balance_threshold,
  high_balance_threshold
) VALUES (
  $1, $2, $3
)
ON CONFLICT (account_id)
DO UPDATE SET
  low_balance_threshold = EXCLUDED.low_balance_threshold,
  high_balance_threshold = EXCLUDED.high_balance_threshold,
  updated_at = now()
RETURNING *;

-- name: GetAccountAlert :one
SELECT * FROM account_alerts
WHERE account_id = $1 LIMIT 1;
