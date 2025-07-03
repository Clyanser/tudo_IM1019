package group_models

import (
	"tudo_IM1019/common/models"
	"tudo_IM1019/common/models/ctype"
)

type GroupModel struct {
	models.Model
	Title                string                     `gorm:"32" json:"title"`      // 群名
	Abstract             string                     `gorm:"128" json:"abstract"`  // 简介
	Avatar               string                     `gorm:"256" json:"avatar"`    // 群头像
	Creator              uint                       `json:"creator"`              // 群主
	IsSearch             bool                       `json:"isSearch"`             // 是否可以被搜索
	Verification         int8                       `json:"verification"`         // 群验证 0 不允许任何人添加  1 允许任何人添加  2 需要验证消息 3 需要回答问题  4  需要正确回答问题
	VerificationQuestion ctype.VerificationQuestion `json:"verificationQuestion"` // 验证问题  为3和4的时候需要
	IsInvite             bool                       `json:"isInvite"`             // 是否可邀请好友
	IsTemporarySession   bool                       `json:"isTemporarySession"`   // 是否开启临时会话
	IsProhibition        bool                       `json:"isProhibition"`        // 是否开启全员禁言
	Size                 int                        `json:"size"`                 // 群规模  20  100 200 1000 2000
}
