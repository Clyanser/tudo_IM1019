package logic

import (
	"context"
	"errors"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"time"
	"tudo_IM1019/tudoIM_auth/auth_api/internal/svc"
	"tudo_IM1019/utils/jwts"
)

type LogoutLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LogoutLogic) Logout(token string) (resp string, err error) {
	if token == "" {
		err = errors.New("token is empty")
		return
	}
	//验证token
	payload, err := jwts.ParseToken(token, l.svcCtx.Config.Auth.AccessSecret)
	if err != nil {
		err = errors.New("invalid token")
		return
	}
	//计算过期时间
	now := time.Now()
	expiration := payload.ExpiresAt.Time.Sub(now)

	key := fmt.Sprintf("logout_%s", token)

	l.svcCtx.RDB.SetNX(key, "", expiration)

	resp = "注销成功"
	return
}
