-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS "public"."oauth_token" (
  "id" text NOT NULL,
  "provider" text NOT NULL,
  "user_id" text NOT NULL,
  "user_name" text NOT NULL,
  "token_type" text NOT NULL,
  "access_token" text NOT NULL,
  "refresh_token" text NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "scope" text NOT NULL,
  "v" int NOT NULL,
  "cat" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "uat" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,

  PRIMARY KEY ("id"),
  UNIQUE ("provider", "user_id")
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "public"."oauth_token";
-- +goose StatementEnd
