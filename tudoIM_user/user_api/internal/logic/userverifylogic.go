package logic

import (
	"context"
	"errors"
	"tudo_IM1019/tudoIM_user/user_models"

	"tudo_IM1019/tudoIM_user/user_api/internal/svc"
	"tudo_IM1019/tudoIM_user/user_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserVerifyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserVerifyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserVerifyLogic {
	return &UserVerifyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserVerifyLogic) UserVerify(req *types.UserVerifyRequest) (resp *types.UserVerifyResponse, err error) {
	//判断是否已经是好友
	var friend user_models.FriendModel
	if friend.IsFriend(l.svcCtx.DB, req.UserID, req.FriendID) {
		return nil, errors.New("你们已经是好友了~")
	}

	var userConf user_models.UserConfModel
	err = l.svcCtx.DB.Take(&userConf, "user_id = ?", req.FriendID).Error
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	resp = new(types.UserVerifyResponse)
	// 将被添加用户的好友验证方式返回给客户端，用于指导前端展示对应界面
	resp.FriendVerification = userConf.FriendVerification
	switch userConf.FriendVerification {
	case 0: // 不允许任何人添加
	case 1: // 允许任何人添加
		// 直接成为好友
		// 先往验证表里面加一条记录，然后通过
	case 2: // 需要验证问题
	case 3, 4: // 需要正确回答问题
		if userConf.VerificationQuestion != nil {
			resp.VerificationQuestion = types.VerificationQuestion{
				Problem1: userConf.VerificationQuestion.Problem1,
				Problem2: userConf.VerificationQuestion.Problem1,
				Problem3: userConf.VerificationQuestion.Problem1,
			}
		}
	default:

	}

	return
}
