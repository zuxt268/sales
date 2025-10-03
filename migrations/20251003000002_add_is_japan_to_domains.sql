-- +migrate Up
ALTER TABLE domains ADD COLUMN is_japan BOOLEAN NOT NULL DEFAULT false AFTER can_view;

-- +migrate Down
ALTER TABLE domains DROP COLUMN is_japan;