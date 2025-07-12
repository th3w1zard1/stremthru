-- +goose Up
-- +goose StatementBegin
ALTER TABLE "public"."torrent_info" ADD COLUMN "release_types" text NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE "public"."torrent_info" DROP COLUMN "release_types";
-- +goose StatementEnd
