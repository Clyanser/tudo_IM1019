package logic

import (
	"context"
	"errors"
	"tudo_IM1019/tudoIM_user/user_models"

	"tudo_IM1019/tudoIM_user/user_api/internal/svc"
	"tudo_IM1019/tudoIM_user/user_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type FriendNoticeUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFriendNoticeUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendNoticeUpdateLogic {
	return &FriendNoticeUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FriendNoticeUpdateLogic) FriendNoticeUpdate(req *types.FriendNoticeUpdateRequest) (resp *types.FriendNoticeUpdateResponse, err error) {
	var friend user_models.FriendModel
	// 验证好友关系
	if !friend.IsFriend(l.svcCtx.DB, req.UserID, req.FriendID) {
		return nil, errors.New("TA 还不是你的好友")
	}
	if friend.SendUserID == req.UserID {
		//我是发起方
		//懒惰机制
		if friend.Notice == req.Notice {
			return
		}
		l.svcCtx.DB.Model(&friend).Update("notice", req.Notice)
	}
	if friend.RevUserID == req.UserID {
		//我是接收方
		//懒惰机制
		if friend.Notice == req.Notice {
			return
		}
		l.svcCtx.DB.Model(&friend).Update("notice", req.Notice)
	}

	return
}
