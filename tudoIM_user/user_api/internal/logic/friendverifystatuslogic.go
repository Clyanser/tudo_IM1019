package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"tudo_IM1019/common/models/ctype"
	"tudo_IM1019/tudoIM_chat/chat_rpc/chat"
	"tudo_IM1019/tudoIM_user/user_models"

	"tudo_IM1019/tudoIM_user/user_api/internal/svc"
	"tudo_IM1019/tudoIM_user/user_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type FriendVerifyStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFriendVerifyStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendVerifyStatusLogic {
	return &FriendVerifyStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FriendVerifyStatusLogic) FriendVerifyStatus(req *types.FriendVerifyStatusRequest) (resp *types.FriendVerifyStatusResponse, err error) {
	var friendVerify user_models.FriendVerifyModel
	// 我要操作状态，我自己得是接收方
	err = l.svcCtx.DB.Take(&friendVerify, "id = ? and rev_user_id = ?", req.VerifyID, req.UserID).Error
	if err != nil {
		return nil, errors.New("验证记录不存在")
	}
	if friendVerify.Status != 0 {
		return nil, errors.New("不可更改状态")
	}

	switch req.Status {
	case 1: // 同意
		friendVerify.Status = 1
		// 往好友表里面加
		l.svcCtx.DB.Create(&user_models.FriendModel{
			SendUserID: friendVerify.SendUserID,
			RevUserID:  friendVerify.RevUserID,
		})

		//给对方发一个消息
		msg := ctype.Msg{
			Type: 1,
			TextMsg: &ctype.TextMsg{
				Content: "我们已经是好友了,开始聊天吧",
			},
		}
		byteData, err := json.Marshal(msg)
		if err != nil {
			logx.Errorf("JSON marshal failed: %v", err) // 打印错误信息
			return nil, fmt.Errorf("无法序列化消息: %v", err)
		}
		logx.Infof("Serialized message: %s", string(byteData)) // 打印序列化后的消息

		_, err = l.svcCtx.ChatRpc.UserChat(context.Background(), &chat.UserChatRequest{
			SendUserId: uint32(req.UserID),
			RevUserId:  uint32(friendVerify.SendUserID),
			Msg:        byteData,
			SystemMsg:  nil,
		})
		if err != nil {
			return nil, err
		}
	case 2: // 拒绝
		friendVerify.Status = 2
	case 3: // 忽略
		friendVerify.Status = 3
	case 4: // 删除
		// 一条验证记录，是两个人看的
		l.svcCtx.DB.Delete(&friendVerify)
		return nil, nil
	}
	l.svcCtx.DB.Save(&friendVerify)
	return
}
