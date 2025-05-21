-- +goose Up
-- +goose StatementBegin
ALTER TABLE "public"."mdblist_item" RENAME TO "mdblist_item_old";

CREATE TABLE IF NOT EXISTS "public"."mdblist_item" (
    "imdb_id" text NOT NULL,
    "adult" boolean NOT NULL DEFAULT false,
    "title" text NOT NULL,
    "poster" text NOT NULL,
    "language" text NOT NULL,
    "mediatype" text NOT NULL,
    "release_year" int NOT NULL,
    "spoken_language" text NOT NULL,
    "uat" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY ("imdb_id")
);

INSERT INTO "public"."mdblist_item" ("imdb_id", "adult", "title", "poster", "language", "mediatype", "release_year", "spoken_language", "uat")
SELECT "imdb_id", "adult", "title", "poster", "language", "mediatype", "release_year", "spoken_language", "uat" FROM "public"."mdblist_item_old";

CREATE TABLE IF NOT EXISTS "public"."imdb_title_map" (
    "imdb" text NOT NULL,
    "tmdb" text NOT NULL DEFAULT '',
    "tvdb" text NOT NULL DEFAULT '',
    "trakt" text NOT NULL DEFAULT '',
    "uat" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY ("imdb")
);

INSERT INTO "public"."imdb_title_map" ("imdb", "tmdb", "tvdb", "uat")
SELECT "imdb_id", "id"::text, CASE WHEN "tvdb_id" = 0 THEN '' ELSE "tvdb_id"::text END, "uat"
FROM "public"."mdblist_item_old";

DROP TABLE "public"."mdblist_item_old";

ALTER TABLE "public"."mdblist_list_item" RENAME TO "mdblist_list_item_old";

CREATE TABLE IF NOT EXISTS "public"."mdblist_list_item" (
    "list_id" int NOT NULL,
    "item_id" text NOT NULL,
    "rank" int NOT NULL DEFAULT 0,

    PRIMARY KEY ("list_id", "item_id")
);

INSERT INTO "public"."mdblist_list_item" ("list_id", "item_id")
SELECT mlio."list_id", itm."imdb" FROM "public"."mdblist_list_item_old" mlio JOIN "public"."imdb_title_map" itm ON itm."tmdb" = mlio."item_id"::text;

DROP TABLE "public"."mdblist_list_item_old";

ALTER TABLE "public"."mdblist_item_genre" RENAME TO "mdblist_item_genre_old";

CREATE TABLE IF NOT EXISTS "public"."mdblist_item_genre" (
    "item_id" varchar NOT NULL,
    "genre" varchar NOT NULL,

    PRIMARY KEY ("item_id", "genre")
);

INSERT INTO "public"."mdblist_item_genre" ("item_id", "genre")
SELECT itm."imdb", migo."genre" FROM "public"."mdblist_item_genre_old" migo JOIN "public"."imdb_title_map" itm ON itm."tmdb" = migo."item_id"::text;

DROP TABLE "public"."mdblist_item_genre_old";
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
