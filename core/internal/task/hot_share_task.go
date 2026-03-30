package task

import (
	"cloud_disk/core/internal/cache"
	"cloud_disk/core/internal/models"
	"context"
	"log"
	"sync"
	"time"
	"xorm.io/xorm"
)

// HotShareTask 热门分享统计任务
type HotShareTask struct {
	engine     *xorm.Engine
	shareCache *cache.ShareCache // redis
	stopChan   chan struct{}
	stopOnce   sync.Once
}

func NewHotShareTask(engine *xorm.Engine, shareCache *cache.ShareCache) *HotShareTask {
	return &HotShareTask{
		engine:     engine,
		shareCache: shareCache,
		stopChan:   make(chan struct{}),
	}
}

// Start 启动定时任务
func (t *HotShareTask) Start() {
	// 立即执行一次
	t.updateHotShares()

	// 每 10 分钟执行一次
	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				t.updateHotShares()
			case <-t.stopChan:
				ticker.Stop()
				return
			}
		}
	}()

	log.Println("热门分享统计任务已启动，每 10 分钟更新一次")
}

// Stop 停止定时任务
func (t *HotShareTask) Stop() {
	t.stopOnce.Do(func() {
		close(t.stopChan)
		log.Println("热门分享统计任务已停止")
	})
}

// updateHotShares 更新热门分享列表
func (t *HotShareTask) updateHotShares() {
	ctx := context.Background()

	identities, err := t.shareCache.GetDailyTopShares(ctx, 100)
	if err != nil {
		log.Printf("获取日榜热门分享失败: %v", err)
		return
	}

	if len(identities) == 0 {
		log.Println("日榜为空，从数据库查询今天活跃的分享初始化热榜")

		var shares []models.ShareBasic
		today := time.Now().Format("2006-01-02")

		err := t.engine.
			Where("deleted_at IS NULL").
			And("DATE(updated_at) = ?", today).
			OrderBy("click_num DESC").
			Limit(100).
			Find(&shares)

		if err != nil {
			log.Printf("从数据库查询今天活跃分享失败: %v", err)
			return
		}

		if len(shares) == 0 {
			log.Println("今天暂无活跃分享，跳过初始化")
			return
		}

		identities = make([]string, 0, len(shares))
		for _, share := range shares {
			identities = append(identities, share.Identity)
		}

		log.Printf("从数据库初始化热榜，共 %d 个今天活跃的分享", len(identities))
	}

	err = t.shareCache.SetHotShareList(ctx, identities)
	if err != nil {
		log.Printf("保存热门分享列表到 Redis 失败: %v", err)
		return
	}

	log.Printf("日榜热门分享列表已更新，共 %d 个分享", len(identities))
}
