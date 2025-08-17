package logic

import (
	"context"
	"errors"
	"tudo_IM1019/common/list_query"
	"tudo_IM1019/common/models"
	"tudo_IM1019/common/models/ctype"
	"tudo_IM1019/tudoIM_chat/chat_models"
	"tudo_IM1019/tudoIM_user/user_rpc/types/user_rpc"
	"tudo_IM1019/utils"

	"tudo_IM1019/tudoIM_chat/chat_api/internal/svc"
	"tudo_IM1019/tudoIM_chat/chat_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChatHistoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChatHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChatHistoryLogic {
	return &ChatHistoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

type UserInfo struct {
	ID       uint   `json:"id"`
	NickName string `json:"nickName"`
	Avatar   string `json:"avatar"`
}

type ChatHistory struct {
	ID        uint             `json:"id"`
	SendUser  UserInfo         `json:"sendUser"`
	RevUser   UserInfo         `json:"revUser"`
	IsMe      bool             `json:"isMe"`       // 哪条消息是我发的
	CreatedAt string           `json:"created_at"` // 消息时间
	Msg       ctype.Msg        `json:"msg"`
	SystemMsg *ctype.SystemMsg `json:"systemMsg"`
}
type ChatHistoryResponse struct {
	List  []ChatHistory `json:"List"`
	Count int64         `json:"count"`
}

func (l *ChatHistoryLogic) ChatHistory(req *types.ChatHistoryRequest) (resp *ChatHistoryResponse, err error) {
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

	chatList, count, _ := list_query.ListQuery(l.svcCtx.DB, chat_models.ChatModel{}, list_query.Option{
		PageInfo: models.PageInfo{
			Page:  req.Page,
			Limit: req.Limit,
			Sort:  "created_at desc", //按照时间降序
		},
		Debug: true, //方便调试与问题定位
		Where: l.svcCtx.DB.Where("((send_user_id = ? and rev_user_id = ?) or (send_user_id = ? and rev_user_id = ?)) and id not in (select chat_id from user_chat_delete_models where user_id = ?)", req.UserID, req.FriendID, req.FriendID, req.UserID, req.UserID),
	})
	var userIDList []uint32
	for _, model := range chatList {
		userIDList = append(userIDList, uint32(model.SendUserID))
		userIDList = append(userIDList, uint32(model.RevUserID))
	}

	//去重
	userIDList = utils.Unique(userIDList)
	//去掉用户服务的rpc方法，获取用户信息 {用户id:{用户信息}}
	response, err := l.svcCtx.UserRpc.UserListInfo(context.Background(), &user_rpc.UserListInfoRequest{
		UserIdList: userIDList,
	})
	if err != nil {
		logx.Error(err)
		return nil, errors.New("用户服务错误")
	}

	var list = make([]ChatHistory, 0)
	for _, model := range chatList {

		sendUser := UserInfo{
			ID:       model.SendUserID,
			NickName: response.UserInfo[uint32(model.SendUserID)].NickName,
			Avatar:   response.UserInfo[uint32(model.SendUserID)].Avatar,
		}
		revUser := UserInfo{
			ID:       model.RevUserID,
			NickName: response.UserInfo[uint32(model.RevUserID)].NickName,
			Avatar:   response.UserInfo[uint32(model.RevUserID)].Avatar,
		}
		list = append(list, ChatHistory{
			ID:        uint(model.ID),
			CreatedAt: model.CreatedAt.String(),
			SendUser:  sendUser,
			RevUser:   revUser,
			Msg:       model.Msg,
			SystemMsg: model.SystemMsg,
		})
	}
	resp = &ChatHistoryResponse{
		Count: count,
		List:  list,
	}
	return
}
