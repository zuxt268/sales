
-- +migrate Up
INSERT INTO tasks (name, description, status)
VALUES
    ('check_open', 'サイトが開くか確認します', 1),
    ('check_japan_owner_id', '日本のサイトか確認します', 1),
    ('check_comp_info', '企業情報を抽出します', 1),
    ('kick_gpt', '代表者名、企業名などを解析します', 1);

-- +migrate Down
DELETE FROM tasks where 1;