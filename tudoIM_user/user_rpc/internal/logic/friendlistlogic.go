package logic

import (
	"context"
	"tudo_IM1019/common/list_query"
	"tudo_IM1019/common/models"
	"tudo_IM1019/tudoIM_user/user_models"

	"tudo_IM1019/tudoIM_user/user_rpc/internal/svc"
	"tudo_IM1019/tudoIM_user/user_rpc/types/user_rpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type FriendListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFriendListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendListLogic {
	return &FriendListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *FriendListLogic) FriendList(in *user_rpc.FriendListRequest) (*user_rpc.FriendListResponse, error) {
	// todo: add your logic here and delete this line
	friends, _, _ := list_query.ListQuery(l.svcCtx.DB, user_models.FriendModel{}, list_query.Option{
		PageInfo: models.PageInfo{
			Limit: -1,
		},
		Preload: []string{"SendUserModel", "RevUserModel"},
	})
	var list []*user_rpc.FriendInfo
	for _, friend := range friends {
		info := user_rpc.FriendInfo{}
		if friend.SendUserID == uint(in.User) {
			//	我是发起方
			info = user_rpc.FriendInfo{
				UserId:   uint32(friend.RevUserID),
				Avatar:   friend.RevUserModel.Avatar,
				NickName: friend.RevUserModel.Nickname,
			}
		}
		if friend.RevUserID == uint(in.User) {
			//	我是接收方
			info = user_rpc.FriendInfo{
				UserId:   uint32(friend.SendUserID),
				Avatar:   friend.SendUserModel.Avatar,
				NickName: friend.SendUserModel.Nickname,
			}
		}
		list = append(list, &info)
	}
	return &user_rpc.FriendListResponse{FriendList: list}, nil
}
