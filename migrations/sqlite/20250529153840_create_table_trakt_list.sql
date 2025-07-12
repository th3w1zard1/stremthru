-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS `trakt_list` (
    `id` varchar NOT NULL,
    `user_id` varchar NOT NULL,
    `user_name` varchar NOT NULL,
    `name` varchar NOT NULL,
    `slug` varchar NOT NULL,
    `description` varchar NOT NULL DEFAULT '',
    `private` bool NOT NULL DEFAULT false,
    `likes` int NOT NULL DEFAULT 0,
    `uat` datetime NOT NULL DEFAULT (unixepoch()),

    PRIMARY KEY (`id`)
);

CREATE TABLE IF NOT EXISTS `trakt_item` (
    `id` int NOT NULL,
    `type` varchar NOT NULL,
    `title` varchar NOT NULl,
    `year` int NOT NULL,
    `overview` varchar NOT NULl,
    `runtime` int NOT NULL,
    `poster` varchar NOT NULL,
    `fanart` varchar NOT NULL,
    `trailer` varchar NOT NULL,
    `rating` int NOT NULL,
    `mpa_rating` varchar NOT NULL,
    `uat` datetime NOT NULL DEFAULT (unixepoch()),

    PRIMARY KEY (`id`, `type`)
);

CREATE TABLE IF NOT EXISTS `trakt_item_genre` (
  `item_id` int NOT NULL,
  `item_type` varchar NOT NULL,
  `genre` varchar NOT NULL,

  PRIMARY KEY (`item_id`, `item_type`, `genre`)
);

CREATE TABLE IF NOT EXISTS `trakt_list_item` (
  `list_id` varchar NOT NULL,
  `item_id` int NOT NULL,
  `item_type` varchar NOT NULL,
  `idx` int NOT NULL,

  PRIMARY KEY (`list_id`, `item_id`, `item_type`)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS `trakt_list_item`;
DROP TABLE IF EXISTS `trakt_item_genre`;
DROP TABLE IF EXISTS `trakt_item`;
DROP TABLE IF EXISTS `trakt_list`;
-- +goose StatementEnd
