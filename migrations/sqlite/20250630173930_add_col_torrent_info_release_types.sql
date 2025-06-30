-- +goose Up
-- +goose StatementBegin
ALTER TABLE `torrent_info` ADD COLUMN `release_types` varchar NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `torrent_info` DROP COLUMN `release_types`;
-- +goose StatementEnd
