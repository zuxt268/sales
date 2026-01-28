-- +migrate Up
ALTER TABLE homstas
    ADD COLUMN server VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'サーバー' AFTER domain,
    ADD COLUMN mail_usage VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'メール使用量' AFTER disc_usage;

-- +migrate Down
ALTER TABLE homstas
    DROP COLUMN server,
    DROP COLUMN mail_usage;