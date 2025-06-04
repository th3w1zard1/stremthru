-- +goose Up
-- +goose StatementBegin
ALTER TABLE "public"."mdblist_list" ALTER COLUMN "id" TYPE text USING "id"::text;
ALTER TABLE "public"."mdblist_list_item" ALTER COLUMN "list_id" TYPE text USING "list_id"::text;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE "public"."mdblist_list_item" ALTER COLUMN "list_id" TYPE int USING "list_id"::int;
ALTER TABLE "public"."mdblist_list" ALTER COLUMN "id" TYPE int USING "id"::int;
-- +goose StatementEnd
