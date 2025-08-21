package file_models

import (
	"github.com/google/uuid"
	"tudo_IM1019/common/models"
)

type FileModel struct {
	models.Model
	FileUid  uuid.UUID `json:"fileUid"`  // 文件唯一ID /api/file/{uuid}
	UserID   uint      `json:"userID"`   //用户ID
	FileName string    `json:"fileName"` //文件名称
	Size     int64     `json:"size"`     //文件大小
	Path     string    `json:"path"`     //文件的实际路径
	Hash     string    `json:"hash"`     //文件hash
}

func (file *FileModel) WebPath() string {
	return "/api/file/" + file.FileUid.String()
}
