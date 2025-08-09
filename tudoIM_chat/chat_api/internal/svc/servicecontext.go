package svc

import (
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/gorm"
	"tudo_IM1019/core"
	"tudo_IM1019/tudoIM_chat/chat_api/internal/config"
	"tudo_IM1019/tudoIM_user/user_rpc/types/user_rpc"
	"tudo_IM1019/tudoIM_user/user_rpc/users"
)

type ServiceContext struct {
	Config  config.Config
	DB      *gorm.DB
	UserRpc user_rpc.UsersClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:  c,
		DB:      core.InitGorm(c.Mysql.Dsn),
		UserRpc: users.NewUsers(zrpc.MustNewClient(c.UserRpc)),
	}
}
