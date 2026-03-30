package models

import "time"

// RepositoryPool 表示网盘中的“物理文件资源池”。
//
// 这里记录的是去重后的真实文件：
// - hash / size 用来判断是否可以秒传
// - path / objectKey 用来找到 OSS 上真实对象
// 多个用户的 user_repository 可以共同引用同一条 repository_pool。
type RepositoryPool struct {
	Id        int
	Identity  string
	Hash      string
	Name      string
	Ext       string
	Size      int64
	Path      string
	ObjectKey string
	CreatedAt time.Time `xorm:"created"`
	UpdatedAt time.Time `xorm:"updated"`
	DeletedAt time.Time `xorm:"deleted"`
}

func (RepositoryPool) TableName() string {
	return "repository_pool"
}
