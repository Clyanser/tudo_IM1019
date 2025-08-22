package logic

import (
	"context"
	"errors"
	"strings"
	"tudo_IM1019/tudoIM_file/file_models"

	"tudo_IM1019/tudoIM_file/file_rpc/internal/svc"
	"tudo_IM1019/tudoIM_file/file_rpc/types/file_rpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type FileInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFileInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FileInfoLogic {
	return &FileInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *FileInfoLogic) FileInfo(in *file_rpc.FileInfoRequest) (*file_rpc.FileInfoResponse, error) {
	// 判断文件是否存在
	var fileModel file_models.FileModel
	logx.Infof("file_id = %s", in.FileId)
	err := l.svcCtx.DB.Take(&fileModel, "file_uid= ?", in.FileId).Error
	if err != nil {
		return nil, errors.New("file not found")
	}
	var suffix string
	nameList := strings.Split(fileModel.FileName, ".")
	if len(nameList) > 1 {
		suffix = nameList[len(nameList)-1]
	}
	return &file_rpc.FileInfoResponse{
		FileName: fileModel.FileName,
		FileHash: fileModel.Hash,
		FileSize: fileModel.Size,
		FileType: suffix,
	}, nil
}
