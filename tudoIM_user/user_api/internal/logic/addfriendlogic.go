package logic

import (
	"context"
	"errors"
	"tudo_IM1019/common/models/ctype"
	"tudo_IM1019/tudoIM_user/user_models"

	"tudo_IM1019/tudoIM_user/user_api/internal/svc"
	"tudo_IM1019/tudoIM_user/user_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddFriendLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddFriendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddFriendLogic {
	return &AddFriendLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddFriendLogic) AddFriend(req *types.AddFriendRequest) (resp *types.AddFriendResponse, err error) {
	//验证是否已是好友
	var friend user_models.FriendModel
	if friend.IsFriend(l.svcCtx.DB, req.UserID, req.FriendID) {
		return nil, errors.New("你们已经是好友了~")
	}
	//查询被添加人的，添加好友设置
	var userConf user_models.UserConfModel
	err = l.svcCtx.DB.Take(&userConf, "user_id = ?", req.FriendID).Error
	if err != nil {
		return nil, errors.New("用户不存在！")
	}
	resp = new(types.AddFriendResponse)
	var verifyModel = user_models.FriendVerifyModel{
		SendUserID:         req.UserID,
		RevUserID:          req.FriendID,
		AdditionalMessages: req.Verify,
	}
	switch userConf.FriendVerification {
	case 0:
		return nil, errors.New("该用户不允许任何人添加")
	case 1: //允许任何人添加
		//	直接成为好友
		//  先往验证表里面加一条记录然后通过
		verifyModel.Status = 1
		var userFriend = user_models.FriendModel{
			SendUserID: req.UserID,
			RevUserID:  req.FriendID,
		}
		l.svcCtx.DB.Create(&userFriend)
	case 2: //需要验证消息
	case 3: //需要回答问题
		if req.VerificationQuestion != nil {
			verifyModel.VerificationQuestion = &ctype.VerificationQuestion{
				Problem1: req.VerificationQuestion.Problem1,
				Problem2: req.VerificationQuestion.Problem2,
				Problem3: req.VerificationQuestion.Problem3,
				Answer1:  req.VerificationQuestion.Answer1,
				Answer2:  req.VerificationQuestion.Answer2,
				Answer3:  req.VerificationQuestion.Answer3,
			}
		}
	case 4: // 需要正确回答问题
		// 前置判断
		if req.VerificationQuestion != nil && userConf.VerificationQuestion != nil {
			//	考虑到有三种情况(1,2,3)
			//验证问题逻辑
			var count int
			if userConf.VerificationQuestion.Answer1 != nil && req.VerificationQuestion.Answer1 != nil {
				if *userConf.VerificationQuestion.Answer1 == *req.VerificationQuestion.Answer1 {
					count += 1
				}
			}
			if userConf.VerificationQuestion.Answer2 != nil && req.VerificationQuestion.Answer2 != nil {
				if *userConf.VerificationQuestion.Answer2 == *req.VerificationQuestion.Answer2 {
					count += 1
				}
			}
			if userConf.VerificationQuestion.Answer3 != nil && req.VerificationQuestion.Answer3 != nil {
				if *userConf.VerificationQuestion.Answer3 == *req.VerificationQuestion.Answer3 {
					count += 1
				}
			}
			if count != userConf.ProblemCount() {
				return nil, errors.New("答案错误")
			}
			//加好友逻辑
			verifyModel.Status = 1
			verifyModel.VerificationQuestion = userConf.VerificationQuestion
			// 加好友
			var userFriend = user_models.FriendModel{
				SendUserID: req.UserID,
				RevUserID:  req.FriendID,
			}
			l.svcCtx.DB.Create(&userFriend)
		}
	default:
		return nil, errors.New("非法的验证参数")
	}
	err = l.svcCtx.DB.Create(&verifyModel).Error
	if err != nil {
		logx.Error(err)
		return nil, errors.New("添加好友失败")
	}
	return
}
