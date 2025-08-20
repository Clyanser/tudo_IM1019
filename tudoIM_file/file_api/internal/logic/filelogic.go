package logic

import (
	"context"

	"tudo_IM1019/tudoIM_file/file_api/internal/svc"
	"tudo_IM1019/tudoIM_file/file_api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type FileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FileLogic {
	return &FileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FileLogic) File(req *types.FileRequest) (resp *types.FileResponse, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.FileResponse)
	return
}
