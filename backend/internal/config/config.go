package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	Mongo MongoDBConfig // mongo 配置
	// AI 服务配置
	AI    AIConfig
	Qiniu QiniuConfig // 七牛云配置
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
	PrometheusURL string `json:",optional"`      // Prometheus URL（MCP 模式需要）
}

type QiniuConfig struct {
	AccessKey    string // 七牛云 Access Key
	SecretKey    string // 七牛云 Secret Key
	Bucket       string // 七牛云存储桶名称
	DownloadHost string
}
