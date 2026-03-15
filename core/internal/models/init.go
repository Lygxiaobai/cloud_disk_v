package models

import (
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"xorm.io/xorm"
)

// 初始化数据库操作对象
func Init(dataSource string) *xorm.Engine {
	var err error
	Engine, err := xorm.NewEngine("mysql", dataSource)
	if err != nil {
		log.Printf("Xorm NewEngine err:%v", err)
		return nil
	}
	return Engine
}

func InitRedis(addr string) *redis.Client {
	Rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return Rdb
}
