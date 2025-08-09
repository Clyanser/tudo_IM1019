package logic

import (
	"context"
	"errors"
	"tudo_IM1019/tudoIM_user/user_models"

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
	//验证是否已是好友
	var friend user_models.FriendModel
	if friend.IsFriend(l.svcCtx.DB, req.UserID, req.FriendID) {
		return nil, errors.New("你们还不是好友~")
	}
	err = l.svcCtx.DB.Where("send_user_id = ? AND rev_user_id = ?",
		req.UserID, req.FriendID).Delete(&user_models.FriendModel{}).Error

	if err != nil {
		logx.Errorf("FriendDelete err:%v", err)
		return nil, err
	}
	return
}
