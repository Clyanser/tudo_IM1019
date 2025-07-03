package group_models

import "tudo_IM1019/common/models"

// GroupMemberModel 群成员表
type GroupMemberModel struct {
	models.Model
	GroupID         uint       `json:"groupID"`                       // 群id
	GroupModel      GroupModel `gorm:"foreignKey:GroupID" json:"-"`   // 群
	UserID          uint       `json:"userID"`                        // 用户id
	MemberNickname  string     `gorm:"size:32" json:"memberNickname"` // 群成员昵称
	Role            int        `json:"role"`                          // 1 群主 2 管理员  3 普通成员
	ProhibitionTime *int       `json:"prohibitionTime"`               // 禁言时间 单位分钟
}
