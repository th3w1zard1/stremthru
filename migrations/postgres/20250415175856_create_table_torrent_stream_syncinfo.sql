-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS "public"."torrent_stream_syncinfo" (
    "sid" text NOT NULL,
    "pulled_at" timestamptz,
    "pushed_at" timestamptz,

    PRIMARY KEY ("sid")
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "public"."torrent_stream_syncinfo";
-- +goose StatementEnd
