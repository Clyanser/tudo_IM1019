package logic

import (
	"context"
	"errors"
	"tudo_IM1019/tudoIM_auth/auth_models"
	"tudo_IM1019/utils/jwts"
	"tudo_IM1019/utils/pwd"

	"tudo_IM1019/tudoIM_auth/auth_api/internal/svc"
	"tudo_IM1019/tudoIM_auth/auth_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginRequest) (resp *types.LoginResponse, err error) {
	var user auth_models.UserModel
	err = l.svcCtx.DB.Take(&user, "id = ?", req.UserName).Error
	if err != nil {
		err = errors.New("用户名或密码错误")
		return
	}

	if !pwd.CheckPwd(user.Pwd, req.Password) {
		err = errors.New("用户名或密码错误")
		return
	}
	//生成token
	token, err := jwts.GenToken(jwts.JwtPayLoad{
		UserID:   user.ID,
		Nickname: user.Nickname,
		Role:     user.Role,
	}, l.svcCtx.Config.Auth.AccessSecret, l.svcCtx.Config.Auth.AccessExpire)
	if err != nil {
		logx.Error(err)
		err = errors.New("服务内部错误")
		return
	}
	//发放token
	resp = &types.LoginResponse{
		Token: token,
	}

	return resp, nil
}
