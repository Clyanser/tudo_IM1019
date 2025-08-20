package chat_models

import "tudo_IM1019/common/models"

// TopUserModel 置顶用户表
type TopUserModel struct {
	models.Model
	UserID    uint `json:"user_id"`
	TopUserID uint `json:"top_user_id"`
}
