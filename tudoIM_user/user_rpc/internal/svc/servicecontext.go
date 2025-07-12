package svc

import (
	"gorm.io/gorm"
	"tudo_IM1019/core"
	"tudo_IM1019/tudoIM_user/user_rpc/internal/config"
)

type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
		DB:     core.InitGorm(c.Mysql.Dsn),
	}
}
