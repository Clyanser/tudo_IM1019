package svc

import (
	"github.com/zeromicro/go-zero/zrpc"
	"tudo_IM1019/tudoIM_file/file_api/internal/config"
	"tudo_IM1019/tudoIM_user/user_rpc/types/user_rpc"
	"tudo_IM1019/tudoIM_user/user_rpc/users"
)

type ServiceContext struct {
	Config  config.Config
	UserRpc user_rpc.UsersClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:  c,
		UserRpc: users.NewUsers(zrpc.MustNewClient(c.UserRpc)),
	}
}
