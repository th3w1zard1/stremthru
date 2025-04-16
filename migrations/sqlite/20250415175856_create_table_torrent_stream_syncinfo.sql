-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS `torrent_stream_syncinfo` (
    `sid` varchar NOT NULL,
    `pulled_at` datetime,
    `pushed_at` datetime,

    PRIMARY KEY (`sid`)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS `torrent_stream_syncinfo`;
-- +goose StatementEnd
