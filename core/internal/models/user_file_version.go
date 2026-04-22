package models

import "time"

// UserFileVersion 记录“逻辑文件”的历史版本快照。
//
// 这里的 file_identity 指向当前用户看到的那条 user_repository 记录，
// repository_identity 则指向某个时间点对应的物理文件版本。
type UserFileVersion struct {
	Id                 int
	Identity           string
	UserIdentity       string
	FileIdentity       string
	RepositoryIdentity string
	Name               string
	Ext                string
	Size               int64
	Hash               string
	Action             string
	CreatedAt          time.Time `xorm:"created"`
}

func (UserFileVersion) TableName() string {
	return "user_file_version"
}
