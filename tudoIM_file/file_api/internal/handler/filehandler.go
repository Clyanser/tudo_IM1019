package handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"tudo_IM1019/tudoIM_file/file_api/internal/logic"
	"tudo_IM1019/tudoIM_file/file_api/internal/svc"
	"tudo_IM1019/tudoIM_file/file_api/internal/types"
	"tudo_IM1019/tudoIM_file/file_models"
	"tudo_IM1019/tudoIM_user/user_rpc/types/user_rpc"
	"tudo_IM1019/utils"

	"tudo_IM1019/common/response"
)

func FileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FileRequest
		if err := httpx.ParseHeaders(r, &req); err != nil {
			response.Response(r, w, nil, err)
			return
		}

		file, fileHead, err := r.FormFile("file")
		if err != nil {
			response.Response(r, w, nil, err)
			return
		}

		//文件上传黑名单
		nameList := strings.Split(fileHead.Filename, ".")
		var suffix string
		if len(nameList) > 1 {
			suffix = nameList[len(nameList)-1]
		}
		if utils.InList(svcCtx.Config.BlackList, suffix) {
			response.Response(r, w, nil, errors.New("文件非法~"))
			return
		}
		//先判断hash 是否重复
		l := logic.NewFileLogic(r.Context(), svcCtx)
		resp, err := l.File(&req)
		FileData, err := io.ReadAll(file)
		if err != nil {
			response.Response(r, w, nil, err)
			return
		}
		fileHash := utils.MD5(FileData)

		var fileModel file_models.FileModel
		err = svcCtx.DB.Take(&fileModel, "hash = ?", fileHash).Error
		if err == nil {
			resp.Src = fileModel.WebPath()
			logx.Infof("文件 %s Hash重复", fileHead.Filename)
			response.Response(r, w, nil, err)
			return
		}

		//文件重名
		//在文件保存之前，先去读文件列表，对比两图片的hash 若一样则直接使用原存在的图片不做写操作 若不一样就把最新的这个重命名以下 {old_name}_xxxx.{suffix}

		//先去拿用户信息
		useResponse, err := svcCtx.UserRpc.UserListInfo(context.Background(), &user_rpc.UserListInfoRequest{
			UserIdList: []uint32{uint32(req.UserID)},
		})
		if err != nil {
			response.Response(r, w, nil, err)
			return
		}
		//拼接路径
		dir1 := fmt.Sprintf("%d_%s", req.UserID, useResponse.UserInfo[uint32(req.UserID)].NickName)

		dirPath := path.Join(svcCtx.Config.UploadDir, "file", dir1)
		_, err = os.ReadDir(dirPath)
		if err != nil {
			err := os.MkdirAll(dirPath, 0666)
			if err != nil {
				return
			}
		}
		newFileModel := file_models.FileModel{
			UserID:   req.UserID,
			FileName: fileHead.Filename,
			Size:     fileHead.Size,
			Hash:     fileHash,
			FileUid:  uuid.New(),
		}
		newFileModel.Path = path.Join(dirPath, fmt.Sprintf("%s.%s", newFileModel.FileUid, suffix))

		err = os.WriteFile(newFileModel.Path, FileData, 0666)
		if err != nil {
			response.Response(r, w, nil, err)
			return
		}
		//文件信息入库
		err = svcCtx.DB.Create(&newFileModel).Error
		if err != nil {
			logx.Error(err)
			response.Response(r, w, resp, err)
			return
		}
		resp.Src = newFileModel.WebPath()
		response.Response(r, w, resp, err)
	}
}
