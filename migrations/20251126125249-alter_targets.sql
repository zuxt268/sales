-- +migrate Up
ALTER TABLE targets
ADD COLUMN current_page INT NOT NULL DEFAULT 1 COMMENT '次に取得するページ番号',
ADD COLUMN last_fetched_at DATETIME NULL DEFAULT NULL COMMENT '最後に取得した日時',
ADD COLUMN last_full_scan_at DATETIME NULL DEFAULT NULL COMMENT 'フルスキャン完了日時';

-- +migrate Down
ALTER TABLE targets
DROP COLUMN current_page,
DROP COLUMN last_fetched_at,
DROP COLUMN last_full_scan_at;