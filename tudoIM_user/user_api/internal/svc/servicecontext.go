package svc

import (
	"github.com/go-redis/redis"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/gorm"
	"tudo_IM1019/core"
	"tudo_IM1019/tudoIM_user/user_api/internal/config"
	"tudo_IM1019/tudoIM_user/user_rpc/types/user_rpc"
	"tudo_IM1019/tudoIM_user/user_rpc/users"
)

type ServiceContext struct {
	Config  config.Config
	DB      *gorm.DB
	RDB     *redis.Client
	UserRpc user_rpc.UsersClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:  c,
		DB:      core.InitGorm(c.Mysql.Dsn),
		RDB:     core.InitRedis(c.Redis.Addr, c.Redis.Pwd, c.Redis.DB),
		UserRpc: users.NewUsers(zrpc.MustNewClient(c.UserRpc)),
	}
}
