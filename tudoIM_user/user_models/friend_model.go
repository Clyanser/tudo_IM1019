package user_models

import (
	"gorm.io/gorm"
	"tudo_IM1019/common/models"
)

type FriendModel struct {
	models.Model
	SendUserID    uint      `json:"sendUserID"`                     // 发起验证方
	SendUserModel UserModel `gorm:"foreignKey:SendUserID" json:"-"` // 发起验证方
	RevUserID     uint      `json:"revUserID"`                      // 接受验证方
	RevUserModel  UserModel `gorm:"foreignKey:RevUserID" json:"-"`  // 接受验证方
	Notice        string    `gorm:"size:128" json:"notice"`         // 备注
}

func (f *FriendModel) IsFriend(db *gorm.DB, a, b uint) bool {
	err := db.Take(&f, "(send_user_id = ? and rev_user_id = ?) or (rev_user_id = ? and send_user_id = ?)", a, b, b, a).Error
	if err != nil {
		return false
	}
	return true
}
