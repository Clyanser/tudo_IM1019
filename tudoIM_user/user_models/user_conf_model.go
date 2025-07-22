package user_models

import (
	"tudo_IM1019/common/models"
	"tudo_IM1019/common/models/ctype"
)

type UserConfModel struct {
	models.Model
	UserID               uint                        `json:"user_id"`
	UserModel            UserModel                   `gorm:"foreignKey:UserID" json:"-"`
	RecallMessage        *string                     `json:"recall_message"`        //撤回消息的回复内容
	FriendOnline         bool                        `json:"friend_online"`         //好友上线提醒
	Sound                bool                        `json:"sound"`                 //提醒声音
	SecureLink           bool                        `json:"source_link"`           //安全衔接
	SavePwd              bool                        `json:"save_pwd"`              //保存密码
	SearchUser           int8                        `json:"search_user"`           //别人查找到你的方式 0-不允许别人查找到我 ，1-通过用户查找到我  2-可以通过昵称搜索到我
	FriendVerification   int8                        `json:"friend_verification"`   //好友验证 0-不允许任何人添加 1-允许任何人添加 2-需要验证消息 3-需要回答问题 4-需要正确回答问题
	VerificationQuestion *ctype.VerificationQuestion `json:"verification_question"` //好友验证问题类型
	IsOnline             bool                        `json:"is_online"`
}
