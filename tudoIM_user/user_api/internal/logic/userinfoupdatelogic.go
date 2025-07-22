package logic

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"tudo_IM1019/common/models/ctype"
	"tudo_IM1019/tudoIM_user/user_models"
	"tudo_IM1019/utils/maps"

	"tudo_IM1019/tudoIM_user/user_api/internal/svc"
	"tudo_IM1019/tudoIM_user/user_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserInfoUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserInfoUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserInfoUpdateLogic {
	return &UserInfoUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserInfoUpdateLogic) UserInfoUpdate(req *types.UserInfoUpdateRequest) (resp *types.UserInfoUpdateResponse, err error) {
	// 打印用户ID
	fmt.Println("用户ID：", req.UserID)

	// 处理 user 相关字段
	userMaps := maps.RefToMap(*req, "user")
	if len(userMaps) > 0 {
		var user user_models.UserModel
		if err := l.svcCtx.DB.Take(&user, req.UserID).Error; err != nil {
			return nil, errors.New("用户不存在")
		}

		if err := l.svcCtx.DB.Model(&user).Updates(userMaps).Error; err != nil {
			logx.Error("用户信息更新失败：", err)
			return nil, errors.New("用户信息更新失败")
		}
	}

	// 处理 user_conf 相关字段
	userConfMaps := maps.RefToMap(*req, "user_conf")
	if len(userConfMaps) > 0 {
		var userConf user_models.UserConfModel
		if err := l.svcCtx.DB.Where("user_id = ?", req.UserID).Take(&userConf).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("用户配置不存在")
			}
			logx.Error("查询用户配置失败：", err)
			return nil, errors.New("用户配置查询失败")
		}

		// 构造更新对象
		updateData := user_models.UserConfModel{}

		// 单独处理 verification_question 字段
		if v, ok := userConfMaps["verification_question"]; ok {
			delete(userConfMaps, "verification_question")

			// 解析 JSON map 到结构体
			rawMap, ok := v.(map[string]interface{})
			if !ok {
				return nil, errors.New("verification_question 格式错误")
			}

			var question ctype.VerificationQuestion
			if err := maps.MapToStruct(rawMap, &question); err != nil {
				logx.Error("解析 verification_question 失败：", err)
				return nil, errors.New("无法解析 verification_question")
			}

			updateData.VerificationQuestion = &question
		}

		// 映射其他字段
		if err := maps.MapToStruct(userConfMaps, &updateData); err != nil {
			logx.Error("字段映射失败：", err)
			return nil, errors.New("字段映射失败")
		}

		// 一次性更新所有字段
		if err := l.svcCtx.DB.Model(&userConf).Updates(updateData).Error; err != nil {
			logx.Error("用户配置更新失败：", err)
			return nil, errors.New("用户配置更新失败")
		}
	}

	//return &types.UserInfoUpdateResponse{
	//	Message: "更新成功",
	//}, nil
	return
}
