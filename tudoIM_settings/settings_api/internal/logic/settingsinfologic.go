package logic

import (
	"context"

	"tudo_IM1019/tudoIM_settings/settings_api/internal/svc"
	"tudo_IM1019/tudoIM_settings/settings_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SettingsInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSettingsInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SettingsInfoLogic {
	return &SettingsInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SettingsInfoLogic) SettingsInfo(req *types.SettingsInfoRequest) (resp *types.SettingsInfoResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
