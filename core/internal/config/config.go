// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	Mysql struct {
		DataSource string
	}
	Redis struct {
		Addr     string
		Password string
		DB       int
	}
	RabbitMQ struct {
		URL           string
		EmailQueue    string
		LogExchange   string
		LocalLogQueue string
		ESLogQueue    string
	}
	Elasticsearch struct {
		Addresses   []string
		Username    string
		Password    string
		IndexPrefix string
	}
	Casbin struct {
		ModelPath  string
		PolicyPath string
	}
	OSS       OSSConfig
	JWT       JWTConfig
	Mail      MailConfig
	CORS      CORSConfig
	Upload    UploadConfig
	RateLimit RateLimitConfig
}

type OSSConfig struct {
	Region               string
	Bucket               string
	Endpoint             string
	AccessKeyId          string
	AccessKeySecret      string
	RoleArn              string
	ExternalID           string
	UploadBaseDir        string
	StsDurationSeconds   int64
	PreviewExpireSeconds int64
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessExpire  int
	RefreshExpire int
}

type MailConfig struct {
	From       string
	Host       string
	Username   string
	Password   string
	ServerName string
	CodeExpire int
	CodeLen    int
}

type CORSConfig struct {
	AllowedOrigins []string
}

type UploadConfig struct {
	MaxSize           int64
	BlockedExtensions []string
}

type RateLimitConfig struct {
	LoginPerMinute         int
	RegisterPerHour        int
	MailCodePerEmailMinute int
	MailCodePerEmailHour   int
	LoginLockThreshold     int
	LoginLockMinutes       int
}
