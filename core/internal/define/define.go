package define

import "github.com/golang-jwt/jwt/v4"

// UserClaim JWT Claims
type UserClaim struct {
	ID       int
	Identity string
	Name     string
	Role     string
	jwt.StandardClaims
}

// 分页参数默认值
var Page = 1
var PageSize = 10

// 分享文件过期时间 s
var FileExpireTime = 60 * 60 * 24

// 点击次数
var DefaultClickNum = 0

// 旧版 OSS 直传路径用到的静态值
// 新版 STS 上传已改由 config.OSS 注入，这两个留作兼容旧版 helper.FileUpload
var Region = "cn-hangzhou"
var BucketName = "lvxiaobai"
