package models

import "time"

//公共存储池

type RepositoryPool struct {
	Id        int
	Identity  string
	Hash      string
	Name      string
	Ext       string //文件扩展名
	Size      int64
	Path      string
	CreatedAt time.Time `xorm:"created"`
	UpdatedAt time.Time `xorm:"updated"`
	DeletedAt time.Time `xorm:"deleted"`
}

func (RepositoryPool) TableName() string {
	return "repository_pool"
}
