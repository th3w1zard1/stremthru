-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS `oauth_token` (
  `id` varchar NOT NULL,
  `provider` varchar NOT NULL,
  `user_id` varchar NOT NULL,
  `user_name` varchar NOT NULL,
  `token_type` varchar NOT NULL,
  `access_token` varchar NOT NULL,
  `refresh_token` varchar NOT NULL,
  `expires_at` datetime NOT NULL,
  `scope` varchar NOT NULL,
  `v` int NOT NULL,
  `cat` datetime NOT NULL DEFAULT (unixepoch()),
  `uat` datetime NOT NULL DEFAULT (unixepoch()),

  PRIMARY KEY (`id`),
  UNIQUE (`provider`, `user_id`)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS `oauth_token`;
-- +goose StatementEnd
