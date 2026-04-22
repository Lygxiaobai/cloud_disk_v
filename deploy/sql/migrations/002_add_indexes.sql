CREATE INDEX idx_ur_user_identity ON user_repository(user_identity);
CREATE INDEX idx_ur_parent_id ON user_repository(parent_id);
CREATE INDEX idx_ur_identity ON user_repository(identity);
CREATE INDEX idx_ur_deleted_at ON user_repository(deleted_at);
CREATE INDEX idx_ur_user_parent_deleted ON user_repository(user_identity, parent_id, deleted_at);
CREATE INDEX idx_ur_user_deleted ON user_repository(user_identity, deleted_at);

CREATE UNIQUE INDEX idx_rp_identity ON repository_pool(identity);
CREATE INDEX idx_rp_hash_size ON repository_pool(hash, size);

CREATE UNIQUE INDEX idx_sb_identity ON share_basic(identity);
CREATE INDEX idx_sb_user_identity ON share_basic(user_identity);

CREATE UNIQUE INDEX idx_us_identity ON upload_session(identity);
CREATE INDEX idx_us_user_status ON upload_session(user_identity, status);

CREATE INDEX idx_ufv_file_identity ON user_file_version(file_identity);

CREATE UNIQUE INDEX idx_ub_identity ON user_basic(identity);
CREATE UNIQUE INDEX idx_ub_name ON user_basic(name);
CREATE UNIQUE INDEX idx_ub_email ON user_basic(email);
