package logic

import (
	"context"
	"errors"
	"tudo_IM1019/tudoIM_chat/chat_models"
	"tudo_IM1019/tudoIM_user/user_rpc/types/user_rpc"

	"tudo_IM1019/tudoIM_chat/chat_api/internal/svc"
	"tudo_IM1019/tudoIM_chat/chat_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserTopLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserTopLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserTopLogic {
	return &UserTopLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserTopLogic) UserTop(req *types.UserTopRequest) (resp *types.UserTopResponse, err error) {
	if req.UserID != req.FriendID { //实现自己给自己置顶
		//是否是好友
		res, err := l.svcCtx.UserRpc.IsFriend(context.Background(), &user_rpc.IsFriendRequest{
			User1: uint32(req.UserID),
			User2: uint32(req.FriendID),
		})
		if err != nil {
			return nil, err
		}
		if !res.IsFriend {
			return nil, errors.New("你们还不是好友~")
		}
	}

	var userTop chat_models.TopUserModel
	err1 := l.svcCtx.DB.Take(&userTop, "user_id = ? and top_user_id = ?", req.UserID, req.FriendID).Error
	if err1 != nil {
		//没有置顶、
		l.svcCtx.DB.Create(&chat_models.TopUserModel{
			UserID:    req.UserID,
			TopUserID: req.FriendID,
		})
		return
	}
	//已经有置顶了
	//会发生gorm bug （待探究）
	//l.svcCtx.DB.Debug().Model(chat_models.TopUserModel{}).Delete("user_id = ? and top_user_id = ?", req.UserID, req.FriendID)
	l.svcCtx.DB.Delete(&userTop)
	return
}
