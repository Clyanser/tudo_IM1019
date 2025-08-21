package handler

import (
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
	"tudo_IM1019/utils"

	"tudo_IM1019/common/response"
)

func ImageHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ImageRequest
		if err := httpx.Parse(r, &req); err != nil {
			response.Response(r, w, nil, err)
			return
		}
		imageType := r.FormValue("imageType")
		switch imageType {
		case "avatar", "group_avatar", "chat":
		case "":
			response.Response(r, w, nil, errors.New("image type is empty"))
			return
		default:
			response.Response(r, w, nil, errors.New("imageType只能为avatar、group_avatar、chat"))
			return
		}
		file, fileHead, err := r.FormFile("image")
		if err != nil {
			response.Response(r, w, nil, err)
			return
		}

		// 文件大小限制
		imageSize := float64(fileHead.Size) / float64(1024) / float64(1024)
		if imageSize > svcCtx.Config.FileSize {
			response.Response(r, w, nil, fmt.Errorf("图片大小超过限制~ 最大只能上传%.2f大小的图片", svcCtx.Config.FileSize))
			return
		}
		//文件后缀白名单
		nameList := strings.Split(fileHead.Filename, ".")
		var suffix string
		if len(nameList) > 1 {
			suffix = nameList[len(nameList)-1]
		}
		if !utils.InList(svcCtx.Config.WhiteList, suffix) {
			response.Response(r, w, nil, errors.New("图片非法~"))
			return
		}
		//先去算hash
		imageData, _ := io.ReadAll(file)
		imageHash := utils.MD5(imageData)
		l := logic.NewImageLogic(r.Context(), svcCtx)
		resp, err := l.Image(&req)
		if err != nil {
			response.Response(r, w, nil, err)
		}
		var fileModel file_models.FileModel
		err = svcCtx.DB.Take(&fileModel, "hash = ?", imageHash).Error
		if err == nil {
			//找到了，返回之前的那个文件的hash
			resp.Url = fileModel.WebPath()
			logx.Infof("文件 %s Hash重复", fileHead.Filename)
			response.Response(r, w, nil, err)
			return
		}

		//拼路径 /uploads/imageType/{uuid}.{后缀 }
		dirPath := path.Join(svcCtx.Config.UploadDir, imageType)
		_, err = os.ReadDir(dirPath)
		if err != nil {
			err := os.MkdirAll(dirPath, 0666)
			if err != nil {
				return
			}
		}
		fileName := fileHead.Filename
		newFileModel := file_models.FileModel{
			UserID:   req.UserID,
			FileName: fileName,
			Size:     fileHead.Size,
			Hash:     utils.MD5(imageData),
			FileUid:  uuid.New(),
		}
		newFileModel.Path = path.Join(dirPath, fmt.Sprintf("%s.%s", newFileModel.FileUid, suffix))

		err = os.WriteFile(newFileModel.Path, imageData, 0666)
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
		resp.Url = newFileModel.WebPath()
		response.Response(r, w, resp, err)
	}
}

func InDir(dir []os.DirEntry, fileName string) bool {
	for _, entry := range dir {
		if entry.Name() == fileName {
			return true
		}
	}
	return false
}
