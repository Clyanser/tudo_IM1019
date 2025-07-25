package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"tudo_IM1019/utils"
	"tudo_IM1019/utils/random"

	"github.com/zeromicro/go-zero/rest/httpx"
	"tudo_IM1019/tudoIM_file/file_api/internal/logic"
	"tudo_IM1019/tudoIM_file/file_api/internal/svc"
	"tudo_IM1019/tudoIM_file/file_api/internal/types"

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
		//文件重名
		//在文件保存之前，先去读文件列表，对比两图片的hash 若一样则直接使用原存在的图片不做写操作 若不一样就把最新的这个重命名以下 {old_name}_xxxx.{suffix}
		dirPath := path.Join(svcCtx.Config.UploadDir, imageType)
		dir, err := os.ReadDir(dirPath)
		if err != nil {
			err := os.MkdirAll(dirPath, 0666)
			if err != nil {
				return
			}
		}
		filePath := path.Join(svcCtx.Config.UploadDir, imageType, fileHead.Filename)
		imageData, err := io.ReadAll(file)
		fileName := fileHead.Filename
		if err != nil {
			response.Response(r, w, nil, err)
		}

		l := logic.NewImageLogic(r.Context(), svcCtx)
		resp, err := l.Image(&req)
		resp.Url = "/" + filePath

		if InDir(dir, fileHead.Filename) {
			//重名了

			//先读之前的文件
			byteData, _ := os.ReadFile(filePath)
			oleFileHash := utils.MD5(byteData)
			newFileHash := utils.MD5(imageData)
			if newFileHash == oleFileHash {
				//两个文件是一样的
				fmt.Println("each both same file hash")
				response.Response(r, w, resp, nil)
				return
			}
			//两个文件是不一样的
			//改名操作
			var prefix = utils.GetFilePrefix(fileName)
			newPath := fmt.Sprintf("%s_%s.%s", prefix, random.RandStr(4), suffix)
			filePath = path.Join(svcCtx.Config.UploadDir, imageType, newPath)

		}

		err = os.WriteFile(filePath, imageData, 0666)
		if err != nil {
			response.Response(r, w, nil, err)
			return
		}
		resp.Url = "/" + filePath
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
