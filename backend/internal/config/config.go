package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	Mongo MongoDBConfig // mongo 配置
}

type MongoDBConfig struct {
	URL      string // mongo 连接地址
	Database string // mongo 数据库名称
}
