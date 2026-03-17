package cache

import (
	"cloud_disk/core/internal/types"
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

// ShareCache 分享缓存管理
type ShareCache struct {
	rdb *redis.Client
}

func NewShareCache(rdb *redis.Client) *ShareCache {
	return &ShareCache{rdb: rdb}
}

// 缓存 key 定义
const (
	// 热门分享列表（存储前100的identity）
	HotShareListKey = "share:hot:list"
	// 分享详情前缀
	ShareDetailPrefix = "share:detail:"
	// 缓存过期时间
	HotShareExpire    = 15 * time.Minute // 热门列表15分钟过期
	ShareDetailExpire = 1 * time.Hour    // 详情1小时过期   防止用户改文件后redis没有及时更新
)

// ctx的传入可以通过上下文来控制redis操作超过5s自动取消
// SetHotShareList 设置热门分享列表（前100的identity）
func (c *ShareCache) SetHotShareList(ctx context.Context, identities []string) error {
	// 删除旧的列表
	c.rdb.Del(ctx, HotShareListKey)

	// 如果列表为空，直接返回
	if len(identities) == 0 {
		return nil
	}

	// 批量添加到 Redis List
	for _, identity := range identities {
		//直接将list存入redis
		c.rdb.RPush(ctx, HotShareListKey, identity)
	}

	// 设置过期时间
	c.rdb.Expire(ctx, HotShareListKey, HotShareExpire)

	return nil
}

// IsHotShare 判断是否是热门分享
func (c *ShareCache) IsHotShare(ctx context.Context, identity string) bool {
	// 检查是否在热门列表中
	result, err := c.rdb.LRange(ctx, HotShareListKey, 0, -1).Result()
	if err != nil {
		return false
	}

	for _, id := range result {
		if id == identity {
			return true
		}
	}
	return false
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
