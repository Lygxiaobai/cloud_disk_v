package cache

import (
	"cloud_disk/core/internal/types"
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
	"time"
)

// ShareCache 分享缓存管理
type ShareCache struct {
	rdb   *redis.Client
	group singleflight.Group
}

func NewShareCache(rdb *redis.Client) *ShareCache {
	return &ShareCache{rdb: rdb}
}

// 缓存 key 定义
const (
	// 热门分享集合（存储前100的identity）- 使用 SET 代替 List
	HotShareSetKey = "share:hot:set"
	// 分享详情前缀
	ShareDetailPrefix = "share:detail:"
	// 日榜点击数前缀
	DailyClicksPrefix = "share:daily:clicks:"
	// 缓存过期时间
	HotShareExpire    = 15 * time.Minute   // 热门列表15分钟过期
	ShareDetailExpire = 1 * time.Hour      // 详情1小时过期   防止用户改文件后redis没有及时更新
	DailyClicksExpire = 7 * 24 * time.Hour // 日榜数据保留7天
)

// ctx的传入可以通过上下文来控制redis操作超过5s自动取消
// SetHotShareList 设置热门分享集合（前100的identity）- 使用 SET + Pipeline + RENAME 原子操作
func (c *ShareCache) SetHotShareList(ctx context.Context, identities []string) error {
	// 如果列表为空，直接返回
	if len(identities) == 0 {
		return nil
	}

	// 使用临时 key，避免更新过程中的空窗期
	tempKey := HotShareSetKey + ":temp"

	// 使用 Pipeline 批量操作，减少网络往返
	pipe := c.rdb.Pipeline()

	// 删除旧的临时 key（如果存在）
	pipe.Del(ctx, tempKey)

	// 批量添加到临时 SET
	pipe.SAdd(ctx, tempKey, identities)

	// 设置过期时间
	pipe.Expire(ctx, tempKey, HotShareExpire)

	// 执行 Pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}

	// 原子替换：将临时 key 重命名为正式 key
	// RENAME 是原子操作，不会有空窗期
	err = c.rdb.Rename(ctx, tempKey, HotShareSetKey).Err()
	if err != nil {
		// 如果 RENAME 失败，清理临时 key
		c.rdb.Del(ctx, tempKey)
		return err
	}

	return nil
}

// IsHotShare 判断是否是热门分享 - 使用 SISMEMBER，时间复杂度 O(1)
func (c *ShareCache) IsHotShare(ctx context.Context, identity string) bool {
	// 使用 SISMEMBER 直接判断，无需遍历
	exists, err := c.rdb.SIsMember(ctx, HotShareSetKey, identity).Result()
	if err != nil {
		return false
	}
	return exists
}

// SetShareDetail 缓存分享详情
func (c *ShareCache) SetShareDetail(ctx context.Context, identity string, detail *types.ShareFileDetailResponse) error {
	key := ShareDetailPrefix + identity

	// 序列化为 JSON
	data, err := json.Marshal(detail)
	if err != nil {
		return err
	}

	// 存入 Redis
	return c.rdb.Set(ctx, key, data, ShareDetailExpire).Err()
}

// GetShareDetail 获取分享详情
func (c *ShareCache) GetShareDetail(ctx context.Context, identity string) (*types.ShareFileDetailResponse, error) {
	key := ShareDetailPrefix + identity

	// 从 Redis 读取
	data, err := c.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		// 缓存未命中
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// 反序列化
	var detail types.ShareFileDetailResponse
	err = json.Unmarshal([]byte(data), &detail)
	if err != nil {
		return nil, err
	}

	return &detail, nil
}

func (c *ShareCache) LoadShareDetail(ctx context.Context, identity string, loader func() (*types.ShareFileDetailResponse, error)) (*types.ShareFileDetailResponse, error) {
	detail, err := c.GetShareDetail(ctx, identity)
	if err != nil || detail != nil {
		return detail, err
	}

	value, err, _ := c.group.Do(identity, func() (interface{}, error) {
		cached, cacheErr := c.GetShareDetail(ctx, identity)
		if cacheErr != nil || cached != nil {
			return cached, cacheErr
		}

		loaded, loadErr := loader()
		if loadErr != nil || loaded == nil {
			return loaded, loadErr
		}

		if setErr := c.SetShareDetail(ctx, identity, loaded); setErr != nil {
			return nil, setErr
		}

		return loaded, nil
	})
	if err != nil || value == nil {
		return nil, err
	}

	detail, _ = value.(*types.ShareFileDetailResponse)
	return detail, nil
}

// IncrClickNum 增加点击次数（Redis计数）
func (c *ShareCache) IncrClickNum(ctx context.Context, identity string) error {
	key := fmt.Sprintf("share:click:%s", identity)
	return c.rdb.Incr(ctx, key).Err()
}

// GetClickNum 获取Redis中的点击次数
func (c *ShareCache) GetClickNum(ctx context.Context, identity string) (int64, error) {
	key := fmt.Sprintf("share:click:%s", identity)
	return c.rdb.Get(ctx, key).Int64()
}

// DeleteShareDetail 删除分享详情缓存
func (c *ShareCache) DeleteShareDetail(ctx context.Context, identity string) error {
	key := ShareDetailPrefix + identity
	return c.rdb.Del(ctx, key).Err()
}

// getTodayKey 获取今日日榜的 Redis key
func (c *ShareCache) getTodayKey() string {
	today := time.Now().Format("2006-01-02")
	return DailyClicksPrefix + today
}

// IncrDailyClick 增加今日点击数（使用 ZSET）- 限制 ZSET 大小防止内存溢出
func (c *ShareCache) IncrDailyClick(ctx context.Context, identity string) error {
	key := c.getTodayKey()

	// 使用 Pipeline 批量操作
	pipe := c.rdb.Pipeline()

	// 使用 ZINCRBY 增加分数
	pipe.ZIncrBy(ctx, key, 1, identity)

	// 只保留前1000名，删除其余（防止恶意访问导致内存溢出）
	// ZREMRANGEBYRANK key 0 -1001 表示删除排名 0 到倒数第1001的元素
	// 即只保留排名最高的1000个
	pipe.ZRemRangeByRank(ctx, key, 0, -1001)

	// 设置过期时间（7天后自动清理）
	pipe.Expire(ctx, key, DailyClicksExpire)

	// 执行 Pipeline
	_, err := pipe.Exec(ctx)
	return err
}

// GetDailyTopShares 获取今日热榜前 N 名
func (c *ShareCache) GetDailyTopShares(ctx context.Context, limit int) ([]string, error) {
	key := c.getTodayKey()
	// 使用 ZREVRANGE 按分数从高到低获取
	return c.rdb.ZRevRange(ctx, key, 0, int64(limit-1)).Result()
}
