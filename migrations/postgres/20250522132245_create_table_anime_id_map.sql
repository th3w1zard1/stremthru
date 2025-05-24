-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS "public"."anime_id_map" (
    "id" serial NOT NULL PRIMARY KEY,
    "type" text NOT NULL,
    "anidb" text,
    "anilist" text,
    "animeplanet" text,
    "anisearch" text,
    "imdb" text,
    "kitsu" text,
    "livechart" text,
    "mal" text,
    "notifymoe" text,
    "tmdb" text,
    "tvdb" text,
    "uat" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS "anime_id_map_uidx_anidb" ON "public"."anime_id_map" ("anidb");
CREATE UNIQUE INDEX IF NOT EXISTS "anime_id_map_uidx_anilist" ON "public"."anime_id_map" ("anilist");
CREATE UNIQUE INDEX IF NOT EXISTS "anime_id_map_uidx_animeplanet" ON "public"."anime_id_map" ("animeplanet");
CREATE UNIQUE INDEX IF NOT EXISTS "anime_id_map_uidx_anisearch" ON "public"."anime_id_map" ("anisearch");
CREATE UNIQUE INDEX IF NOT EXISTS "anime_id_map_uidx_kitsu" ON "public"."anime_id_map" ("kitsu");
CREATE UNIQUE INDEX IF NOT EXISTS "anime_id_map_uidx_livechart" ON "public"."anime_id_map" ("livechart");
CREATE UNIQUE INDEX IF NOT EXISTS "anime_id_map_uidx_mal" ON "public"."anime_id_map" ("mal");
CREATE UNIQUE INDEX IF NOT EXISTS "anime_id_map_uidx_notifymoe" ON "public"."anime_id_map" ("notifymoe");

CREATE TABLE IF NOT EXISTS "public"."anilist_media" (
    "id" int NOT NULL,
    "type" text NOT NULL,
    "title" text NOT NULL,
    "description" text NOT NULL,
    "banner" text NOT NULL,
    "cover" text NOT NULL,
    "duration" int NOT NULL,
    "is_adult" boolean NOT NULL,
    "start_year" int NOT NULL,
    "uat" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY ("id")
);

CREATE TABLE IF NOT EXISTS "public"."anilist_media_genre" (
    "media_id" int NOT NULL,
    "genre" text NOT NULL,

    PRIMARY KEY ("media_id", "genre")
);

CREATE TABLE IF NOT EXISTS "public"."anilist_list" (
    "id" text NOT NULL,
    "uat" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY ("id")
);

CREATE TABLE IF NOT EXISTS "public"."anilist_list_media" (
    "list_id" text NOT NULL,
    "media_id" int NOT NULL,
    "score" int NOT NULL,

    PRIMARY KEY ("list_id", "media_id")
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "public"."anilist_list_media";
DROP TABLE IF EXISTS "public"."anilist_list";
DROP TABLE IF EXISTS "public"."anilist_media_genre";
DROP TABLE IF EXISTS "public"."anilist_media";
DROP TABLE IF EXISTS "public"."anime_id_map";
-- +goose StatementEnd
