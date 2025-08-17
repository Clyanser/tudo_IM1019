package logic

import (
	"context"
	"strconv"
	"tudo_IM1019/common/list_query"
	"tudo_IM1019/common/models"
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
	//friends, count, err := list_query.ListQuery(l.svcCtx.DB, user_models.FriendModel{}, list_query.Option{
	//	PageInfo: models.PageInfo{
	//		Page:  req.Page,
	//		Limit: req.Limit,
	//	},
	//	Preload: []string{"SendUserModel", "RevUserModel"},
	//})
	//if err != nil {
	//	return nil, err
	//}
	//
	//var list []types.FriendInfoResponse
	//for _, friend := range friends {
	//	info := types.FriendInfoResponse{}
	//	if friend.SendUserID == req.UserID {
	//		//我是发起方
	//		info = types.FriendInfoResponse{
	//			friend.RevUserID,
	//			friend.RevUserModel.Nickname,
	//			friend.RevUserModel.Abstract,
	//			friend.RevUserModel.Avatar,
	//			friend.Notice,
	//		}
	//	}
	//	if friend.RevUserID == req.UserID {
	//		user := user_models.FriendModel{}
	//		err = l.svcCtx.DB.Take(&user, "send_user_id = ? and rev_user_id = ?", friend.SendUserID, req.UserID).Error
	//		if err != nil {
	//			return nil, err
	//		}
	//		//我是接收方
	//		info = types.FriendInfoResponse{
	//			friend.SendUserID,
	//			friend.SendUserModel.Nickname,
	//			friend.SendUserModel.Abstract,
	//			friend.SendUserModel.Avatar,
	//			user.Notice,
	//		}
	//	}
	//
	//	list = append(list, info)
	//}
	//
	//return &types.FriendListResponse{
	//	list,
	//	int(count),
	//}, nil
	// 构建查询条件：只查当前用户作为发起方或接收方的记录
	whereCond := l.svcCtx.DB.Where("send_user_id = ? ", req.UserID)

	// 调用通用分页查询，传入查询条件
	friends, count, err := list_query.ListQuery(l.svcCtx.DB, user_models.FriendModel{}, list_query.Option{
		PageInfo: models.PageInfo{
			Page:  req.Page,
			Limit: req.Limit,
		},
		Where:   whereCond, //  关键：传入 WHERE 条件
		Preload: []string{"SendUserModel", "RevUserModel"},
	})
	if err != nil {
		return nil, err
	}

	//查哪些用户在线
	onlineMap := l.svcCtx.RDB.HGetAll("online").Val()
	var onlineUserMap = map[uint]bool{}
	for key, _ := range onlineMap {
		val, err1 := strconv.Atoi(key)
		if err1 != nil {
			logx.Error(err)
			continue
		}
		onlineUserMap[uint(val)] = true
	}

	// 用于去重：记录已添加的好友 userID
	seen := make(map[uint]bool)
	var list []types.FriendInfoResponse

	for _, friend := range friends {
		var info types.FriendInfoResponse

		// 判断当前用户是发起方还是接收方
		if friend.SendUserID == req.UserID {
			// 我是发起方：对方是 rev_user
			info = types.FriendInfoResponse{
				UserID:   friend.RevUserID,
				Nickname: friend.RevUserModel.Nickname,
				Abstract: friend.RevUserModel.Abstract,
				Avatar:   friend.RevUserModel.Avatar,
				Notice:   friend.Notice, // 我对他的备注
				IsOnline: onlineUserMap[friend.RevUserID],
			}
		} else if friend.RevUserID == req.UserID {
			// 我是接收方：对方是 send_user
			info = types.FriendInfoResponse{
				UserID:   friend.SendUserID,
				Nickname: friend.SendUserModel.Nickname,
				Abstract: friend.SendUserModel.Abstract,
				Avatar:   friend.SendUserModel.Avatar,
				Notice:   friend.Notice, // 注意：仍然是这条记录的 Notice（我对他的备注）
				IsOnline: onlineUserMap[friend.SendUserID],
			}
		} else {
			// 安全兜底：跳过不相关的记录（理论上不会走到这里）
			continue
		}

		// 跳过重复好友
		if seen[info.UserID] {
			continue
		}
		seen[info.UserID] = true

		list = append(list, info)
	}

	return &types.FriendListResponse{
		List:  list,
		Count: int(count),
	}, nil
}
