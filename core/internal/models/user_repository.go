package models

import "time"

type UserRepository struct {
	Id                 int //这个也用于当前用户的层级
	Identity           string
	ParentId           int64
	IsDir              int //是否是文件夹
	UserIdentity       string
	RepositoryIdentity string
	Name               string
	Ext                string    //文件扩展名
	CreatedAt          time.Time `xorm:"created"`
	UpdatedAt          time.Time `xorm:"updated"`
	DeletedAt          time.Time `xorm:"deleted"`
}

func (u UserRepository) TableName() string {
	return "user_repository"
}
