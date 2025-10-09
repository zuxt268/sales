-- +migrate Up
CREATE TABLE domains (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE COMMENT 'ドメイン名',
    can_view BOOLEAN NOT NULL DEFAULT false COMMENT '閲覧可能',
    is_japan BOOLEAN NOT NULL DEFAULT false COMMENT '日本語サイト',
    is_send BOOLEAN NOT NULL DEFAULT false COMMENT '送信済み',
    title VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'サイトタイトル',
    owner_id VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'オーナーID',
    address VARCHAR(500) NOT NULL DEFAULT '' COMMENT '住所',
    phone VARCHAR(50) NOT NULL DEFAULT '' COMMENT '電話番号',
    industry VARCHAR(255) NOT NULL DEFAULT '' COMMENT '業種',
    president VARCHAR(255) NOT NULL DEFAULT '' COMMENT '代表者名',
    company VARCHAR(255) NOT NULL DEFAULT '' COMMENT '会社名',
    is_ssl BOOLEAN NOT NULL DEFAULT false COMMENT 'SSL有効',
    raw_page TEXT NOT NULL COMMENT 'ページ生データ',
    page_num INT NOT NULL DEFAULT 0 COMMENT 'ページ数',
    status VARCHAR(50) NOT NULL DEFAULT 'initialize' COMMENT 'ステータス',
    update_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新日時',
    create_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '作成日時',

    -- インデックス
    INDEX idx_domains_status (status),
    INDEX idx_domains_industry (industry),
    INDEX idx_domains_owner_id (owner_id),
    INDEX idx_domains_is_ssl (is_ssl),
    INDEX idx_domains_can_view (can_view),
    INDEX idx_domains_is_send (is_send),
    INDEX idx_domains_is_japan (is_japan)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='ドメイン情報テーブル';

-- +migrate Down
DROP TABLE IF EXISTS domains;