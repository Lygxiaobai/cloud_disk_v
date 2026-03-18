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
}
