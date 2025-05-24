-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS `anime_id_map` (
    `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT,
    `type` varchar NOT NULL,
    `anidb` varchar,
    `anilist` varchar,
    `animeplanet` varchar,
    `anisearch` varchar,
    `imdb` varchar,
    `kitsu` varchar,
    `livechart` varchar,
    `mal` varchar,
    `notifymoe` varchar,
    `tmdb` varchar,
    `tvdb` varchar,
    `uat` datetime NOT NULL DEFAULT (unixepoch())
);

CREATE UNIQUE INDEX IF NOT EXISTS `anime_id_map_uidx_anidb` ON `anime_id_map` (`anidb`);
CREATE UNIQUE INDEX IF NOT EXISTS `anime_id_map_uidx_anilist` ON `anime_id_map` (`anilist`);
CREATE UNIQUE INDEX IF NOT EXISTS `anime_id_map_uidx_animeplanet` ON `anime_id_map` (`animeplanet`);
CREATE UNIQUE INDEX IF NOT EXISTS `anime_id_map_uidx_anisearch` ON `anime_id_map` (`anisearch`);
CREATE UNIQUE INDEX IF NOT EXISTS `anime_id_map_uidx_kitsu` ON `anime_id_map` (`kitsu`);
CREATE UNIQUE INDEX IF NOT EXISTS `anime_id_map_uidx_livechart` ON `anime_id_map` (`livechart`);
CREATE UNIQUE INDEX IF NOT EXISTS `anime_id_map_uidx_mal` ON `anime_id_map` (`mal`);
CREATE UNIQUE INDEX IF NOT EXISTS `anime_id_map_uidx_notifymoe` ON `anime_id_map` (`notifymoe`);

CREATE TABLE IF NOT EXISTS `anilist_media` (
    `id` int NOT NULL,
    `type` varchar NOT NULL,
    `title` varchar NOT NULL,
    `description` varchar NOT NULL,
    `banner` varchar NOT NULL,
    `cover` varchar NOT NULL,
    `duration` int NOT NULL,
    `is_adult` bool NOT NULL,
    `start_year` int NOT NULL,
    `uat` datetime NOT NULL DEFAULT (unixepoch()),

    PRIMARY KEY (`id`)
);

CREATE TABLE IF NOT EXISTS `anilist_media_genre` (
    `media_id` int NOT NULL,
    `genre` varchar NOT NULL,

    PRIMARY KEY (`media_id`, `genre`)
);

CREATE TABLE IF NOT EXISTS `anilist_list` (
    `id` varchar NOT NULL,
    `uat` datetime NOT NULL DEFAULT (unixepoch()),

    PRIMARY KEY (`id`)
);

CREATE TABLE IF NOT EXISTS `anilist_list_media` (
    `list_id` varchar NOT NULL,
    `media_id` int NOT NULL,
    `score` int NOT NULL,

    PRIMARY KEY (`list_id`, `media_id`)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS `anilist_list_media`;
DROP TABLE IF EXISTS `anilist_list`;
DROP TABLE IF EXISTS `anilist_media_genre`;
DROP TABLE IF EXISTS `anilist_media`;
DROP TABLE IF EXISTS `anime_id_map`;
-- +goose StatementEnd
