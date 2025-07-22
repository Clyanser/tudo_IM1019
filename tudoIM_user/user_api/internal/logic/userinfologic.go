package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"tudo_IM1019/tudoIM_user/user_api/internal/svc"
	"tudo_IM1019/tudoIM_user/user_api/internal/types"
	"tudo_IM1019/tudoIM_user/user_models"
	"tudo_IM1019/tudoIM_user/user_rpc/users"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserInfoLogic {
	return &UserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserInfoLogic) UserInfo(req *types.UserInfoRequest) (resp *types.UserInfoResponse, err error) {
	// todo: add your logic here and delete this line
	res, err := l.svcCtx.UserRpc.UserInfo(context.Background(), &users.UserInfoRequest{
		UserId: uint32(req.UserID),
	})
	if err != nil {
		return nil, err
	}
	fmt.Println(string(res.Data))
	var user user_models.UserModel
	err = json.Unmarshal(res.Data, &user)
	if err != nil {
		logx.Error(err)
		return nil, errors.New("数据错误！")
	}

	//// 防止 UserConfModel 为 nil 导致 panic
	//if user.UserConfModel == nil {
	//	user.UserConfModel = &user_models.UserConfModel{
	//		RecallMessage:       false,
	//		FriendOnline:        false,
	//		Sound:               false,
	//		SecureLink:          false,
	//		SavePwd:             false,
	//		SearchUser:          false,
	//		FriendVerification:  false,
	//		VerificationQuestion: nil,
	//	}
	//}

	resp = &types.UserInfoResponse{
		UserID:        uint(user.ID),
		Nickname:      user.Nickname,
		Abstract:      user.Abstract,
		Avatar:        user.Avatar,
		RecallMessage: user.UserConfModel.RecallMessage,
		FriendOnline:  user.UserConfModel.FriendOnline,
		Sound:         user.UserConfModel.Sound,
		SecureLink:    user.UserConfModel.SecureLink,
		SavePwd:       user.UserConfModel.SavePwd,
		SearchUser:    user.UserConfModel.SearchUser,
		Verification:  user.UserConfModel.FriendVerification,
	}

	if user.UserConfModel.VerificationQuestion != nil {
		resp.VerificationQuestion = &types.VerificationQuestion{
			Problem1: user.UserConfModel.VerificationQuestion.Problem1,
			Problem2: user.UserConfModel.VerificationQuestion.Problem2,
			Problem3: user.UserConfModel.VerificationQuestion.Problem3,
			Answer1:  user.UserConfModel.VerificationQuestion.Answer1,
			Answer2:  user.UserConfModel.VerificationQuestion.Answer2,
			Answer3:  user.UserConfModel.VerificationQuestion.Answer3,
		}
	}
	return resp, nil

}
