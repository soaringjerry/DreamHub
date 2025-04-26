-- 数据库初始化脚本 for DreamHub (v4 - Align with embedding model, requires dropping existing table if dimensions mismatch)

-- 1. 启用必要的扩展
CREATE EXTENSION IF NOT EXISTS vector;
-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 2. 创建 conversation_history 表 (不变)
CREATE TABLE IF NOT EXISTS conversation_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    sender_role VARCHAR(10) NOT NULL CHECK (sender_role IN ('user', 'ai')),
    message_content TEXT NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metadata JSONB
);
CREATE INDEX IF NOT EXISTS idx_conversation_history_user_conv_id_ts
ON conversation_history (user_id, conversation_id, timestamp DESC);

-- 3. 创建 documents 表 (不变)
CREATE TABLE IF NOT EXISTS documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    original_filename VARCHAR(255) NOT NULL,
    stored_path VARCHAR(1024) NOT NULL,
    file_size BIGINT NOT NULL,
    content_type VARCHAR(100),
    upload_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processing_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    processing_task_id UUID,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_documents_user_id_upload_time
ON documents (user_id, upload_time DESC);

-- 4. 创建 tasks 表 (不变)
CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY,
    type VARCHAR(100) NOT NULL,
    payload JSONB,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    user_id VARCHAR(255),
    file_id UUID REFERENCES documents(id) ON DELETE SET NULL,
    original_filename VARCHAR(255),
    progress DOUBLE PRECISION DEFAULT 0.0,
    result JSONB,
    error_message TEXT,
    retry_count INT DEFAULT 0,
    max_retries INT DEFAULT 3,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_tasks_status_created_at ON tasks (status, created_at ASC);
CREATE INDEX IF NOT EXISTS idx_tasks_user_id ON tasks (user_id);
CREATE INDEX IF NOT EXISTS idx_tasks_file_id ON tasks (file_id);

-- 5. 处理 langchain_pg_embedding 表 (向量存储)
-- 5.1. 安全地删除可能已存在的索引
DROP INDEX IF EXISTS langchain_pg_embedding_ivfflat_idx;
DROP INDEX IF EXISTS idx_gin_embedding_metadata;

-- 5.2. **删除现有表以确保维度正确 (警告：将删除现有数据!)**
DROP TABLE IF EXISTS langchain_pg_embedding;

-- 5.3. 使用正确的维度 (1536 for text-embedding-3-small) 重新创建表
CREATE TABLE langchain_pg_embedding (
    uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    embedding vector(1536) NOT NULL, -- 明确指定与代码中模型匹配的维度!
    document TEXT,
    cmetadata JSONB
);

-- 5.4. 为向量搜索创建索引 (IVFFlat)
CREATE INDEX langchain_pg_embedding_ivfflat_idx -- 使用 CREATE INDEX 而不是 IF NOT EXISTS，因为表是新创建的
ON langchain_pg_embedding
USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100);

-- 5.5. 为元数据过滤创建 GIN 索引
CREATE INDEX idx_gin_embedding_metadata -- 使用 CREATE INDEX 而不是 IF NOT EXISTS
ON langchain_pg_embedding USING GIN (cmetadata jsonb_path_ops);


-- 6. 创建自动更新 updated_at 时间戳的触发器函数 (不变)
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 7. 将触发器应用到需要自动更新 updated_at 的表 (不变)
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'set_timestamp_documents') THEN
    CREATE TRIGGER set_timestamp_documents
    BEFORE UPDATE ON documents FOR EACH ROW EXECUTE FUNCTION trigger_set_timestamp();
  END IF;
END $$;
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'set_timestamp_tasks') THEN
    CREATE TRIGGER set_timestamp_tasks
    BEFORE UPDATE ON tasks FOR EACH ROW EXECUTE FUNCTION trigger_set_timestamp();
  END IF;
END $$;

-- 提示：数据库初始化完成。
SELECT '数据库初始化脚本 (v4) 执行完毕。';