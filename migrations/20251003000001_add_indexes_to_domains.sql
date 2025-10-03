-- +migrate Up
-- 複合インデックス: よく一緒に使われる条件
CREATE INDEX idx_domains_status ON domains(status);
CREATE INDEX idx_domains_industry ON domains(industry);
CREATE INDEX idx_domains_owner ON domains(owner_id);
CREATE INDEX idx_domains_ssl ON domains(is_ssl);
CREATE INDEX idx_domains_can_view ON domains(can_view);
CREATE INDEX idx_domains_is_japan ON domains(is_japan);
CREATE INDEX idx_domains_is_send ON domains(is_send);
CREATE INDEX idx_domains_owner_id ON domains(owner_id);


-- +migrate Down
DROP INDEX idx_domains_status ON domains;
DROP INDEX idx_domains_industry ON domains;
DROP INDEX idx_domains_owner ON domains;
DROP INDEX idx_domains_ssl ON domains;
DROP INDEX idx_domains_can_view ON domains;
DROP INDEX idx_domains_is_japan ON domains;
DROP INDEX idx_domains_is_send ON domains;
DROP INDEX idx_domains_owner_id ON domains;