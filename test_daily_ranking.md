# 日榜功能测试说明

## 修改内容总结

### 1. Redis Key 设计
- `share:daily:clicks:YYYY-MM-DD` - 日榜 ZSET，记录每日点击数
- `share:hot:list` - 热榜列表 List，存储前100的 identity
- `share:detail:{identity}` - 分享详情缓存

### 2. 核心修改

#### cache/share_cache.go
- 新增 `DailyClicksPrefix` 常量
- 新增 `getTodayKey()` - 生成今日 key
- 新增 `IncrDailyClick()` - 增加今日点击数（ZINCRBY）
- 新增 `GetDailyTopShares()` - 获取今日前100（ZREVRANGE）

#### task/hot_share_task.go
- 修改 `updateHotShares()` - 从日榜 ZSET 获取前100
- 新增冷启动逻辑 - 如果日榜为空，从数据库查询今天活跃的分享（`DATE(updated_at) = 今天`）

#### logic/share-file-detail-logic.go
- 热门分享：异步更新日榜 ZSET + 数据库 click_num
- 非热门分享：同步更新数据库 click_num + 异步更新日榜 ZSET

### 3. 数据流向

```
用户访问分享
    ↓
判断是否在热榜列表（share:hot:list）
    ↓
├─ 是热门：尝试从 Redis 读取详情
│   ├─ 命中：返回缓存
│   └─ 未命中：查数据库 → 写入缓存
│
└─ 不是热门：直接查数据库
    ↓
更新数据库 click_num（持久化）
    ↓
更新日榜 ZSET（今日排名）
    ↓
定时任务每10分钟刷新热榜列表
```

### 4. 冷启动流程

```
系统启动（或新的一天）
    ↓
定时任务立即执行
    ↓
检查日榜 ZSET 是否为空
    ↓
├─ 不为空：从 ZSET 获取前100
│
└─ 为空：从数据库查询
    WHERE DATE(updated_at) = 今天
    ORDER BY click_num DESC
    LIMIT 100
    ↓
保存到 share:hot:list
```

## 测试步骤

### 前置条件
1. 启动 Redis 服务（Docker 或本地）
2. 启动应用服务
3. 确保数据库有分享数据

### 测试场景

#### 场景1：冷启动测试
1. 清空 Redis 所有数据
2. 启动应用
3. 检查日志：应该看到"从数据库查询今天活跃的分享初始化热榜"
4. 检查 Redis：`LRANGE share:hot:list 0 -1` 应该有数据

#### 场景2：日榜更新测试
1. 访问某个分享链接
2. 检查 Redis：`ZSCORE share:daily:clicks:2026-03-18 {identity}` 应该增加
3. 检查数据库：`click_num` 应该增加
4. 等待10分钟后，检查热榜列表是否更新

#### 场景3：缓存命中测试
1. 访问热榜前100的分享
2. 第一次：查数据库，写入缓存
3. 第二次：从 Redis 读取
4. 检查响应时间：第二次应该更快

#### 场景4：跨天测试
1. 修改系统时间到第二天
2. 重启应用
3. 检查 Redis：应该生成新的日榜 key `share:daily:clicks:2026-03-19`
4. 旧的日榜 key 应该在7天后过期

## 验证命令

### Redis 命令
```bash
# 查看热榜列表
LRANGE share:hot:list 0 -1

# 查看今日日榜
ZREVRANGE share:daily:clicks:2026-03-18 0 99 WITHSCORES

# 查看某个分享的今日点击数
ZSCORE share:daily:clicks:2026-03-18 {identity}

# 查看分享详情缓存
GET share:detail:{identity}

# 查看所有日榜 key
KEYS share:daily:clicks:*
```

### 数据库查询
```sql
-- 查看今天更新过的分享
SELECT identity, click_num, updated_at
FROM share_basic
WHERE DATE(updated_at) = CURDATE()
ORDER BY click_num DESC
LIMIT 10;

-- 查看总点击数前10
SELECT identity, click_num, updated_at
FROM share_basic
WHERE deleted_at IS NULL
ORDER BY click_num DESC
LIMIT 10;
```

## 预期结果

1. ✅ 冷启动时能从数据库初始化热榜
2. ✅ 每次访问都更新日榜 ZSET 和数据库
3. ✅ 定时任务每10分钟刷新热榜列表
4. ✅ 热门分享优先从 Redis 读取
5. ✅ 日榜数据7天后自动过期
6. ✅ 每天自动切换到新的日榜 key

## 注意事项

1. 确保数据库的 `updated_at` 字段配置了自动更新
2. Redis 内存足够存储7天的日榜数据
3. 定时任务的执行间隔可以根据实际情况调整
4. 如果访问量特别大，考虑增加 Redis 连接池大小
