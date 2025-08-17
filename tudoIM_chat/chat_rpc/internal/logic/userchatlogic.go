package logic

import (
	"context"
	"encoding/json"
	"tudo_IM1019/common/models/ctype"
	"tudo_IM1019/tudoIM_chat/chat_models"

	"tudo_IM1019/tudoIM_chat/chat_rpc/internal/svc"
	"tudo_IM1019/tudoIM_chat/chat_rpc/types/chat_rpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserChatLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUserChatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserChatLogic {
	return &UserChatLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UserChatLogic) UserChat(in *chat_rpc.UserChatRequest) (*chat_rpc.UserChatResponse, error) {
	//解析msg
	var msg ctype.Msg
	err := json.Unmarshal(in.Msg, &msg)
	if err != nil {
		return nil, err
	}
	var systemMsg *ctype.SystemMsg
	if in.SystemMsg != nil {
		err = json.Unmarshal(in.SystemMsg, systemMsg)
		if err != nil {
			logx.Error(err)
			return nil, err
		}
	}
	chat := chat_models.ChatModel{
		SendUserID: uint(in.SendUserId),
		RevUserID:  uint(in.RevUserId),
		MsgType:    int8(msg.Type),
		Msg:        msg,
		SystemMsg:  systemMsg,
	}
	chat.MsgPreview = chat.MsgPreviewMethod()

	err = l.svcCtx.DB.Create(&chat).Error
	if err != nil {
		logx.Error(err)
		return nil, err
	}

	return &chat_rpc.UserChatResponse{}, nil
}
