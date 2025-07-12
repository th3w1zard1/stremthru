-- +goose Up
-- +goose StatementBegin
ALTER TABLE `mdblist_list` RENAME TO `mdblist_list_old`;

CREATE TABLE `mdblist_list` (
    `id` varchar NOT NULL,
    `user_id` int NOT NULL,
    `user_name` varchar NOT NULL,
    `name` varchar NOT NULL,
    `slug` varchar NOT NULL,
    `description` varchar NOT NULL DEFAULT '',
    `mediatype` varchar NOT NULL,
    `dynamic` bool NOT NULL DEFAULT false,
    `private` bool NOT NULL DEFAULT false,
    `likes` int NOT NULL DEFAULT 0,
    `uat` datetime NOT NULL DEFAULT (unixepoch()),

    PRIMARY KEY (`id`)
);

INSERT INTO `mdblist_list` (`id`, `user_id`, `user_name`, `name`, `slug`, `description`, `mediatype`, `dynamic`, `private`, `likes`, `uat`)
SELECT CAST(`id` AS varchar), `user_id`, `user_name`, `name`, `slug`, `description`, `mediatype`, `dynamic`, `private`, `likes`, `uat` FROM `mdblist_list_old`;

DROP TABLE `mdblist_list_old`;

ALTER TABLE `mdblist_list_item` RENAME TO `mdblist_list_item_old`;

CREATE TABLE IF NOT EXISTS `mdblist_list_item` (
    `list_id` varchar NOT NULL,
    `item_id` varchar NOT NULL,
    `rank` int NOT NULL DEFAULT 0,

    PRIMARY KEY (`list_id`, `item_id`)
);

INSERT INTO `mdblist_list_item` (`list_id`, `item_id`, `rank`)
SELECT CAST(`list_id` AS varchar), `item_id`, `rank` FROM `mdblist_list_item_old`;

DROP TABLE `mdblist_list_item_old`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `mdblist_list_item` RENAME TO `mdblist_list_item_old`;

CREATE TABLE IF NOT EXISTS `mdblist_list_item` (
    `list_id` int NOT NULL,
    `item_id` varchar NOT NULL,
    `rank` int NOT NULL DEFAULT 0,

    PRIMARY KEY (`list_id`, `item_id`)
);

INSERT INTO `mdblist_list_item` (`list_id`, `item_id`, `rank`)
SELECT CAST(`list_id` AS int), `item_id`, `rank` FROM `mdblist_list_item_old`;

DROP TABLE `mdblist_list_item_old`;

ALTER TABLE `mdblist_list` RENAME TO `mdblist_list_old`;

CREATE TABLE `mdblist_list` (
    `id` int NOT NULL,
    `user_id` int NOT NULL,
    `user_name` varchar NOT NULL,
    `name` varchar NOT NULL,
    `slug` varchar NOT NULL,
    `description` varchar NOT NULL DEFAULT '',
    `mediatype` varchar NOT NULL,
    `dynamic` bool NOT NULL DEFAULT false,
    `private` bool NOT NULL DEFAULT false,
    `likes` int NOT NULL DEFAULT 0,
    `uat` datetime NOT NULL DEFAULT (unixepoch()),

    PRIMARY KEY (`id`)
);

INSERT INTO `mdblist_list` (`id`, `user_id`, `user_name`, `name`, `slug`, `description`, `mediatype`, `dynamic`, `private`, `likes`, `uat`)
SELECT CAST(`id` AS int), `user_id`, `user_name`, `name`, `slug`, `description`, `mediatype`, `dynamic`, `private`, `likes`, `uat` FROM `mdblist_list_old`;

DROP TABLE `mdblist_list_old`;
-- +goose StatementEnd
