package models

import "time"

// UploadSession 表示“一次尚未完成或刚完成的上传会话”。
//
// 它和 repository_pool / user_repository 的区别是：
// 1. repository_pool：全局物理文件池
// 2. user_repository：用户看到的逻辑文件
// 3. upload_session：上传中间态，服务于断点续传、暂停继续、STS 续签、complete 收口
type UploadSession struct {
	Id                 int
	Identity           string
	UserIdentity       string
	ParentId           int64
	TargetFileIdentity string
	RepositoryIdentity string
	Name               string
	Ext                string
	Hash               string
	Size               int64
	ObjectKey          string
	Status             string
	CreatedAt          time.Time `xorm:"created"`
	UpdatedAt          time.Time `xorm:"updated"`
	DeletedAt          time.Time `xorm:"deleted"`
}

func (UploadSession) TableName() string {
	return "upload_session"
}
