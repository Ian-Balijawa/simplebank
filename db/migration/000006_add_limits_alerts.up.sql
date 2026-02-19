CREATE TABLE "account_limits" (
  "account_id" bigint PRIMARY KEY,
  "daily_transfer_limit" bigint NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "account_alerts" (
  "account_id" bigint PRIMARY KEY,
  "low_balance_threshold" bigint NOT NULL DEFAULT 0,
  "high_balance_threshold" bigint NOT NULL DEFAULT 0,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

ALTER TABLE "account_limits" ADD FOREIGN KEY ("account_id") REFERENCES "accounts" ("id") ON DELETE CASCADE;

ALTER TABLE "account_alerts" ADD FOREIGN KEY ("account_id") REFERENCES "accounts" ("id") ON DELETE CASCADE;

CREATE INDEX ON "account_limits" ("account_id");
