-- +goose Up
-- +goose StatementBegin
ALTER TABLE `mdblist_item` RENAME TO `mdblist_item_old`;

CREATE TABLE IF NOT EXISTS `mdblist_item` (
    `imdb_id` varchar NOT NULL,
    `adult` bool NOT NULL DEFAULT false,
    `title` varchar NOT NULL,
    `poster` varchar NOT NULL,
    `language` varchar NOT NULL,
    `mediatype` varchar NOT NULL,
    `release_year` int NOT NULL,
    `spoken_language` varchar NOT NULL,
    `uat` datetime NOT NULL DEFAULT (unixepoch()),

    PRIMARY KEY (`imdb_id`)
);

INSERT INTO `mdblist_item` (`imdb_id`, `adult`, `title`, `poster`, `language`, `mediatype`, `release_year`, `spoken_language`, `uat`)
SELECT `imdb_id`, `adult`, `title`, `poster`, `language`, `mediatype`, `release_year`, `spoken_language`, `uat` FROM `mdblist_item_old`;

CREATE TABLE IF NOT EXISTS `imdb_title_map` (
    `imdb` varchar NOT NULL,
    `tmdb` varchar NOT NULL DEFAULT '',
    `tvdb` varchar NOT NULL DEFAULT '',
    `trakt` varchar NOT NULL DEFAULT '',
    `uat` datetime NOT NULL DEFAULT (unixepoch()),

    PRIMARY KEY (`imdb`)
);

INSERT INTO `imdb_title_map` (`imdb`, `tmdb`, `tvdb`, `uat`)
SELECT `imdb_id`, `id`, CASE WHEN `tvdb_id` = 0 THEN '' ELSE `tvdb_id` END, `uat`
FROM `mdblist_item_old`;

DROP TABLE `mdblist_item_old`;

ALTER TABLE `mdblist_list_item` RENAME TO `mdblist_list_item_old`;

CREATE TABLE IF NOT EXISTS `mdblist_list_item` (
    `list_id` int NOT NULL,
    `item_id` varchar NOT NULL,
    `rank` int NOT NULL DEFAULT 0,

    PRIMARY KEY (`list_id`, `item_id`)
);

INSERT INTO `mdblist_list_item` (`list_id`, `item_id`)
SELECT mlio.`list_id`, itm.`imdb` FROM `mdblist_list_item_old` mlio JOIN `imdb_title_map` itm ON itm.`tmdb` = mlio.`item_id`;

DROP TABLE `mdblist_list_item_old`;

ALTER TABLE `mdblist_item_genre` RENAME TO `mdblist_item_genre_old`;

CREATE TABLE IF NOT EXISTS `mdblist_item_genre` (
    `item_id` varchar NOT NULL,
    `genre` varchar NOT NULL,

    PRIMARY KEY (`item_id`, `genre`)
);

INSERT INTO `mdblist_item_genre` (`item_id`, `genre`)
SELECT itm.`imdb`, migo.`genre` FROM `mdblist_item_genre_old` migo JOIN `imdb_title_map` itm ON itm.`tmdb` = migo.`item_id`;

DROP TABLE `mdblist_item_genre_old`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
