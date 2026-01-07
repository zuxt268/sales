-- +migrate Up
CREATE TABLE homstas (
    id INT AUTO_INCREMENT PRIMARY KEY,
    domain VARCHAR(255) NOT NULL COMMENT 'サイト名',
    blog_name TEXT NOT NULL COMMENT 'サイト名',
    path VARCHAR(500) NOT NULL DEFAULT '' UNIQUE COMMENT 'パス',
    site_url VARCHAR(1000) NOT NULL DEFAULT '' COMMENT 'サイトURL',
    description TEXT NOT NULL COMMENT '説明',
    db_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'データベース名',
    users TEXT NOT NULL COMMENT 'ユーザー情報',
    db_usage VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'データベース使用量',
    disc_usage VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'ディスク使用量',
    industry VARCHAR(255) NOT NULL DEFAULT '' COMMENT '業種',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新日時',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '作成日時',

    -- インデックス
    INDEX idx_homstas_path (path)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Homsta情報テーブル';

-- +migrate Down
DROP TABLE IF EXISTS homstas;