package models

type UserBasic struct {
	Id       int
	Identity string
	Name     string
	Password string
	Email    string
	Role     string
}

// 修改表明名
func (u UserBasic) TableName() string {
	return "user_basic"
}
