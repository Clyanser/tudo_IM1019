package logic

import (
	"context"
	"fmt"
	"tudo_IM1019/tudoIM_chat/chat_models"

	"tudo_IM1019/tudoIM_chat/chat_api/internal/svc"
	"tudo_IM1019/tudoIM_chat/chat_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChatDeleteHistoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChatDeleteHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChatDeleteHistoryLogic {
	return &ChatDeleteHistoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChatDeleteHistoryLogic) ChatDeleteHistory(req *types.ChatDeleteHistoryRequest) (resp *types.ChatDeleteHistoryResponse, err error) {
	var chatList []chat_models.ChatModel
	l.svcCtx.DB.Find(&chatList, req.IdList)

	var useDeleteChatList []chat_models.UserChatDeleteModel
	l.svcCtx.DB.Find(&useDeleteChatList, req.IdList)

	chatDeleteMap := map[uint]struct{}{}
	for _, model := range useDeleteChatList {
		chatDeleteMap[model.ChatID] = struct{}{}
	}

	var deleteChatList []chat_models.UserChatDeleteModel

	if len(chatList) > 0 {
		for _, model := range chatList {
			// 不是自己的聊天记录
			if !(model.SendUserID == req.UserID || model.RevUserID == req.UserID) {
				fmt.Println("不是自己的聊天记录", model.ID)
				continue
			}
			// 已经删过的聊天记录
			_, ok := chatDeleteMap[uint(model.ID)]
			if ok {
				fmt.Println("已经删除过了", model.ID)
				continue
			}
			deleteChatList = append(deleteChatList, chat_models.UserChatDeleteModel{
				UserID: req.UserID,
				ChatID: uint(model.ID),
			})
		}
	}
	if len(deleteChatList) > 0 {
		l.svcCtx.DB.Create(&deleteChatList)
	}

	logx.Infof("已删除聊天记录 %d 条", len(deleteChatList))
	return
}
