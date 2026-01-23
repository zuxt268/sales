-- +migrate Up
CREATE TABLE wixes (
    name VARCHAR(255) NOT NULL UNIQUE COMMENT '名前',
    owner_id VARCHAR(255) NOT NULL COMMENT 'オーナーID',

    -- インデックス
    INDEX idx_wixes_owner_id (owner_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Wix情報テーブル';

-- +migrate Down
DROP TABLE IF EXISTS wixes;