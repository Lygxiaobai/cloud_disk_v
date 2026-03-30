package models

import "time"

// UserRepository 表示用户自己的文件目录结构。
//
// 这一张表更像“用户视角下的文件系统”：
// - ParentId 负责层级关系
// - IsDir 判断目录 / 文件
// - IsFavorite 表示收藏状态
// - RepositoryIdentity 指向真实物理文件
type UserRepository struct {
	Id                 int
	Identity           string
	ParentId           int64
	IsDir              int
	IsFavorite         int
	UserIdentity       string
	RepositoryIdentity string
	Name               string
	Ext                string
	CreatedAt          time.Time `xorm:"created"`
	UpdatedAt          time.Time `xorm:"updated"`
	DeletedAt          time.Time `xorm:"deleted"`
}

func (u UserRepository) TableName() string {
	return "user_repository"
}
