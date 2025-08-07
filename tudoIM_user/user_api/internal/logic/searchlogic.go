package logic

import (
	"context"
	"fmt"
	"tudo_IM1019/common/list_query"
	"tudo_IM1019/common/models"
	"tudo_IM1019/tudoIM_user/user_models"

	"tudo_IM1019/tudoIM_user/user_api/internal/svc"
	"tudo_IM1019/tudoIM_user/user_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSearchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchLogic {
	return &SearchLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SearchLogic) Search(req *types.SerachRequest) (resp *types.SerachResponse, err error) {
	//先查所有的用户
	users, count, err := list_query.ListQuery(l.svcCtx.DB, user_models.UserConfModel{}, list_query.Option{
		PageInfo: models.PageInfo{
			Page:  req.Page,
			Limit: req.Limit,
		},
		Preload: []string{"UserModel"},
		Join:    "left join user_models um on um.id = user_conf_models.user_id",
		Where:   l.svcCtx.DB.Where("(user_conf_models.search_user <> 0 or user_conf_models.search_user is not null)   and (user_conf_models.search_user = 1 and um.id = ?)   or (user_conf_models.search_user = 2 and (    um.id = ? or um.nickname like ? ))", req.Key, req.Key, fmt.Sprintf("%%%s%%", req.Key)),
	})
	if err != nil {
		return nil, err
	}
	//获取自己这个用户的好友列表
	var friend user_models.FriendModel
	friends := friend.Friends(l.svcCtx.DB, req.UserID)
	userMap := map[uint]bool{}
	//验证好友关系
	for _, model := range friends {
		if model.SendUserID == req.UserID {
			userMap[model.RevUserID] = true
		} else {
			if model.RevUserID == req.UserID {
				userMap[model.SendUserID] = true
			}
		}

	}

	list := make([]types.SearchInfo, 0)
	for _, uc := range users {
		list = append(list, types.SearchInfo{
			UserID:   uc.UserID,
			Nickname: uc.UserModel.Nickname,
			Abstract: uc.UserModel.Abstract,
			Avatar:   uc.UserModel.Avatar,
			IsFriend: userMap[uc.UserID],
		})
	}

	return &types.SerachResponse{List: list, Count: count}, nil
}

/*
SELECT *
FROM user_conf_models
    left join user_models um on um.id = user_conf_models.user_id
where (user_conf_models.search_user <> 0 or user_conf_models.search_user is not null)
and (user_conf_models.search_user = 1 and um.id = '5')
or (user_conf_models.search_user = 2 and(
        um.id = '5' or um.nickname like '%5%'
    ));
*/
