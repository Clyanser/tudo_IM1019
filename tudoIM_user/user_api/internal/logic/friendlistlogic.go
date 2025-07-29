package logic

import (
	"context"
	"tudo_IM1019/tudoIM_user/user_models"

	"tudo_IM1019/tudoIM_user/user_api/internal/svc"
	"tudo_IM1019/tudoIM_user/user_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type FriendListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFriendListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendListLogic {
	return &FriendListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 获取好友列表
func (l *FriendListLogic) FriendList(req *types.FriendListRequest) (resp *types.FriendListResponse, err error) {
	var count int64
	l.svcCtx.DB.Model(user_models.FriendModel{}).Where("send_user_id = ? or rev_user_id = ?", req.UserID, req.UserID).Count(&count)
	//判断好友关系
	var friends []user_models.FriendModel
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	offset := (req.Page - 1) * req.Limit
	l.svcCtx.DB.Preload("SendUserModel").Preload("RevUserModel").Limit(req.Limit).Offset(offset).Find(&friends, "send_user_id = ? or rev_user_id = ?", req.UserID, req.UserID)
	var list []types.FriendInfoResponse
	for _, friend := range friends {
		info := types.FriendInfoResponse{}
		if friend.SendUserID == req.UserID {
			//我是发起方
			info = types.FriendInfoResponse{
				friend.RevUserID,
				friend.RevUserModel.Nickname,
				friend.RevUserModel.Abstract,
				friend.RevUserModel.Avatar,
				friend.Notice,
			}
		}
		if friend.RevUserID == req.UserID {
			user := user_models.FriendModel{}
			err = l.svcCtx.DB.Take(&user, "send_user_id = ? and rev_user_id = ?", friend.SendUserID, req.UserID).Error
			if err != nil {
				return nil, err
			}
			//我是接收方
			info = types.FriendInfoResponse{
				friend.SendUserID,
				friend.SendUserModel.Nickname,
				friend.SendUserModel.Abstract,
				friend.SendUserModel.Avatar,
				user.Notice,
			}
		}

		list = append(list, info)
	}

	return &types.FriendListResponse{
		list,
		int(count),
	}, nil
}
