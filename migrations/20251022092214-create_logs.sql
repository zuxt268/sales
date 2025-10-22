-- +migrate Up
CREATE TABLE logs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL COMMENT '処理名',
    category VARCHAR(255) NOT NULL DEFAULT 'info' COMMENT 'カテゴリー',
    message TEXT NOT NULL COMMENT 'メッセージ',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '作成日時'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='ログテーブル';

-- +migrate Down
DROP TABLE IF EXISTS logs;