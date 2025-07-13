package logic

import (
	"context"
	"errors"
	"fmt"
	"tudo_IM1019/utils/jwts"

	"tudo_IM1019/tudoIM_auth/auth_api/internal/svc"
	"tudo_IM1019/tudoIM_auth/auth_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AuthenticationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAuthenticationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuthenticationLogic {
	return &AuthenticationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AuthenticationLogic) Authentication(token string) (resp *types.AuthenticationResponse, err error) {
	if token == "" {
		logx.Info("token is empty")
		err = errors.New("token is empty")
		return
	}
	//验证token
	cliams, err := jwts.ParseToken(token, l.svcCtx.Config.Auth.AccessSecret)
	if err != nil {
		logx.Errorf("jwt parse token err: %v", err)
		err = errors.New("jwt parse token err")
		return
	}
	//查看是否已注销
	_, err = l.svcCtx.RDB.Get(fmt.Sprintf("logout_%s", token)).Result()
	if err == nil {
		logx.Errorf("logout_%s is exist", token)
		err = errors.New("logout is exist")
		return
	}

	return &types.AuthenticationResponse{
		UserID: uint(cliams.UserID),
		Role:   int(cliams.Role),
	}, nil
}
