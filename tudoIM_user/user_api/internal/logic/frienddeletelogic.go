package logic

import (
	"context"

	"tudo_IM1019/tudoIM_user/user_api/internal/svc"
	"tudo_IM1019/tudoIM_user/user_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type FriendDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFriendDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendDeleteLogic {
	return &FriendDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FriendDeleteLogic) FriendDelete(req *types.FriendDeleteRequest) (resp *types.FriendDeleteResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
