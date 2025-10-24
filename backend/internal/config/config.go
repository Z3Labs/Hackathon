package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	MongoDB struct {
		URI      string
		Database string
	}
}
