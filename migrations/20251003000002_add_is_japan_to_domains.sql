-- +migrate Up
ALTER TABLE domains ADD COLUMN is_japan BOOLEAN NOT NULL DEFAULT false AFTER can_view;
CREATE INDEX idx_domains_is_japan ON domains(is_japan);

-- +migrate Down
ALTER TABLE domains DROP COLUMN is_japan;
DROP INDEX idx_domains_is_japan ON domains;