package models

import (
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"xorm.io/xorm"
)

// Init 初始化数据库连接并同步当前项目用到的核心表结构。
// 这里使用 Sync2，是因为当前改动以“增量补字段 / 补表”为主，便于开发阶段快速落库。
func Init(dataSource string) *xorm.Engine {
	engine, err := xorm.NewEngine("mysql", dataSource)
	if err != nil {
		log.Printf("xorm new engine failed: %v", err)
		return nil
	}

	if err := engine.Sync2(
		new(UserBasic),
		new(UserRepository),
		new(RepositoryPool),
		new(ShareBasic),
		new(UploadSession),
	); err != nil {
		log.Printf("xorm sync schema failed: %v", err)
		return nil
	}

	return engine
}

// InitRedis 初始化 Redis 客户端。
// 最近文件、验证码、分享缓存等能力都会依赖这个连接。
func InitRedis(addr string, password string, db int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
}
