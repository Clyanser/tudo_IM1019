package logic

import (
	"context"
	"encoding/json"
	"errors"
	"tudo_IM1019/tudoIM_user/user_api/internal/svc"
	"tudo_IM1019/tudoIM_user/user_api/internal/types"
	"tudo_IM1019/tudoIM_user/user_models"
	"tudo_IM1019/tudoIM_user/user_rpc/types/user_rpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type FriendInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFriendInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendInfoLogic {
	return &FriendInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FriendInfoLogic) FriendInfo(req *types.FriendInfoRequest) (resp *types.FriendInfoResponse, err error) {
	//确定你查的用户是自己的好友
	var friend user_models.FriendModel
	if !friend.IsFriend(l.svcCtx.DB, req.UserID, req.FriendID) {
		return nil, errors.New("她还不是你的好友哦！")
	}
	//查用户信息
	res, err := l.svcCtx.UserRpc.UserInfo(context.Background(), &user_rpc.UserInfoRequest{
		UserId: uint32(req.FriendID),
	})
	var user user_models.UserModel
	err = json.Unmarshal(res.Data, &user)
	if err != nil {
		return nil, err
	}
	//初始化friend
	friend = user_models.FriendModel{}
	//查备注
	err = l.svcCtx.DB.Take(&friend, "send_user_id = ? and rev_user_id = ?", req.UserID, req.FriendID).Error
	if err != nil {
		return nil, err
	}

	return &types.FriendInfoResponse{
		UserID:   uint(user.ID),
		Nickname: user.Nickname,
		Abstract: user.Abstract,
		Avatar:   user.Avatar,
		Notice:   friend.Notice,
	}, nil
}
