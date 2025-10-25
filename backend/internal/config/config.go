package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	Mongo  MongoDBConfig // mongo 配置
	AI     AIConfig      // AI 服务配置
	Qiniu  QiniuConfig   // 七牛云配置
}

type MongoDBConfig struct {
	URL      string // mongo 连接地址
	Database string // mongo 数据库名称
}

type AIConfig struct {
	BaseURL       string `json:",optional"` // API 基础 URL，从环境变量读取
	APIKey        string // API 密钥，从环境变量读取
	Model         string `json:",default=gpt-4"` // 模型名称
	Timeout       int    `json:",default=30"`    // 超时时间（秒）
	UseMCP        bool   `json:",default=false"` // 是否使用 MCP 模式
	PrometheusURL string `json:",optional"`      // Prometheus URL（MCP 模式需要）
}

type QiniuConfig struct {
	AccessKey string // 七牛云 AccessKey
	SecretKey string // 七牛云 SecretKey
	Bucket    string `json:",default=niulink-materials"` // 存储桶名称
	Domain    string // 七牛云域名
}
