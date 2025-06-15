# 云端同步功能实现文档

## 概述

DreamHub 云端同步功能允许用户在多个设备之间同步他们的对话历史、结构化记忆、用户配置等数据。该功能采用增量同步策略，支持离线使用和冲突解决。

## 架构设计

### 1. 同步策略
- **增量同步**: 基于时间戳和版本号的变更追踪
- **双向同步**: 支持客户端和服务端双向数据流
- **冲突解决**: Last-Write-Wins (LWW) 策略，用户可选择保留本地或云端版本
- **实时推送**: 计划支持 WebSocket 通知其他设备更新（待实现）

### 2. 数据版本控制
- 为每个需要同步的实体添加了 `version` 和 `sync_at` 字段
- 使用乐观锁处理并发更新
- 自动触发器更新 sync_at 和 version 字段

## 后端实现

### 数据库迁移 (004_add_sync_fields.sql)
```sql
-- 为现有表添加同步相关字段
ALTER TABLE conversations ADD COLUMN version INTEGER DEFAULT 1;
ALTER TABLE conversations ADD COLUMN sync_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE conversations ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE;

-- 创建同步状态表
CREATE TABLE sync_status (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    device_id VARCHAR(255) NOT NULL,
    last_sync_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    sync_version BIGINT DEFAULT 0,
    UNIQUE(user_id, device_id)
);

-- 创建自动更新触发器
CREATE TRIGGER update_conversations_sync_at
BEFORE UPDATE ON conversations
FOR EACH ROW EXECUTE FUNCTION update_sync_at();
```

### 同步 API 端点

#### 1. 获取同步状态
- **端点**: `GET /api/v1/sync/status?device_id={device_id}`
- **功能**: 获取设备的当前同步状态

#### 2. 拉取更新
- **端点**: `POST /api/v1/sync/pull`
- **请求体**:
```json
{
  "device_id": "device-identifier",
  "last_sync_at": "2025-05-01T09:00:00Z",
  "sync_version": 41
}
```
- **响应**: 返回自上次同步以来的所有更改

#### 3. 推送更新
- **端点**: `POST /api/v1/sync/push`
- **请求体**:
```json
{
  "device_id": "device-identifier",
  "sync_version": 42,
  "changes": {
    "conversations": [...],
    "messages": [...],
    "structured_memory": [...],
    "deleted_items": {...}
  }
}
```

#### 4. 解决冲突
- **端点**: `POST /api/v1/sync/conflicts/resolve`
- **功能**: 解决同步冲突

### 核心组件

1. **SyncRepository**: 处理同步相关的数据库操作
   - GetSyncStatus: 获取同步状态
   - GetXXXSince: 获取自某时间以来的变更
   - ApplySyncChanges: 应用同步变更（事务处理）

2. **SyncService**: 处理同步业务逻辑
   - PullChanges: 拉取服务器端变更
   - PushChanges: 推送本地变更
   - DetectConflicts: 检测冲突
   - ResolveConflicts: 解决冲突

3. **SyncHandler**: 处理 HTTP 请求

## 前端实现

### 同步服务 (syncService.ts)
```typescript
class SyncService {
  // 获取同步状态
  async getSyncStatus(): Promise<SyncStatus>
  
  // 拉取变更
  async pullChanges(): Promise<SyncResponse>
  
  // 推送变更
  async pushChanges(changes: SyncChanges): Promise<SyncPushResponse>
  
  // 执行完整同步
  async performSync(localChanges?: SyncChanges): Promise<SyncResult>
  
  // 启动定期同步
  startPeriodicSync(intervalMs: number = 30000): void
}
```

### 同步状态管理 (syncStore.ts)
- 管理同步状态和离线队列
- 处理冲突解决
- 与其他 store 集成

### UI 组件 (SyncStatus.tsx)
- 显示同步状态指示器
- 手动触发同步
- 冲突解决对话框

## 使用流程

1. **初始化**: 用户登录后自动初始化同步，生成设备 ID
2. **自动同步**: 每 30 秒自动执行一次同步
3. **手动同步**: 用户可点击同步按钮手动触发
4. **离线支持**: 离线时变更保存在本地队列，联网后自动同步
5. **冲突处理**: 出现冲突时弹出对话框让用户选择解决方案

## 安全考虑

1. **用户隔离**: 所有同步操作都基于 JWT 认证的用户 ID
2. **设备管理**: 每个设备有唯一标识符，支持跟踪多设备同步
3. **数据完整性**: 使用事务确保同步操作的原子性
4. **版本控制**: 防止数据覆盖和丢失

## 性能优化

1. **增量同步**: 只同步变更的数据，减少传输量
2. **批量操作**: 使用事务批量处理多个变更
3. **索引优化**: 为 sync_at 字段创建索引加速查询
4. **缓存策略**: 本地缓存同步状态减少请求

## 已知限制

1. **冲突检测**: 当前使用简单的 Last-Write-Wins 策略
2. **实时推送**: WebSocket 实时通知尚未实现
3. **大数据量**: 未对超大数据量进行优化
4. **离线时长**: 长时间离线可能导致大量冲突

## 后续改进

1. 实现更智能的冲突解决策略（如三方合并）
2. 添加 WebSocket 支持实时推送
3. 实现选择性同步（如只同步特定对话）
4. 添加同步历史和回滚功能
5. 优化大数据量同步性能