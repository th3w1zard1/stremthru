-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS "public"."trakt_list" (
    "id" text NOT NULL,
    "user_id" text NOT NULL,
    "user_name" text NOT NULL,
    "name" text NOT NULL,
    "slug" text NOT NULL,
    "description" text NOT NULL DEFAULT '',
    "private" boolean NOT NULL DEFAULT false,
    "likes" int NOT NULL DEFAULT 0,
    "uat" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY ("id")
);

CREATE TABLE IF NOT EXISTS "public"."trakt_item" (
    "id" int NOT NULL,
    "type" text NOT NULL,
    "title" text NOT NULl,
    "year" int NOT NULL,
    "overview" text NOT NULl,
    "runtime" int NOT NULL,
    "poster" text NOT NULL,
    "fanart" text NOT NULL,
    "trailer" text NOT NULL,
    "rating" int NOT NULL,
    "mpa_rating" text NOT NULL,
    "uat" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY ("id", "type")
);

CREATE TABLE IF NOT EXISTS "public"."trakt_item_genre" (
  "item_id" int NOT NULL,
  "item_type" text NOT NULL,
  "genre" text NOT NULL,

  PRIMARY KEY ("item_id", "item_type", "genre")
);

CREATE TABLE IF NOT EXISTS "public"."trakt_list_item" (
  "list_id" text NOT NULL,
  "item_id" int NOT NULL,
  "item_type" text NOT NULL,
  "idx" int NOT NULL,

  PRIMARY KEY ("list_id", "item_id", "item_type")
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "public"."trakt_list_item";
DROP TABLE IF EXISTS "public"."trakt_item_genre";
DROP TABLE IF EXISTS "public"."trakt_item";
DROP TABLE IF EXISTS "public"."trakt_list";
-- +goose StatementEnd
