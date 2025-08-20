package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"tudo_IM1019/tudoIM_user/user_rpc/types/user_rpc"
	"tudo_IM1019/utils"
	"tudo_IM1019/utils/random"

	"github.com/zeromicro/go-zero/rest/httpx"
	"tudo_IM1019/tudoIM_file/file_api/internal/logic"
	"tudo_IM1019/tudoIM_file/file_api/internal/svc"
	"tudo_IM1019/tudoIM_file/file_api/internal/types"

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
		dir, err := os.ReadDir(dirPath)
		if err != nil {
			err := os.MkdirAll(dirPath, 0666)
			if err != nil {
				return
			}
		}
		filePath := path.Join(dirPath, fileHead.Filename)
		imageData, err := io.ReadAll(file)
		fileName := fileHead.Filename
		if err != nil {
			response.Response(r, w, nil, err)
		}

		l := logic.NewFileLogic(r.Context(), svcCtx)
		resp, err := l.File(&req)
		resp.Src = "/" + filePath

		if InDir(dir, fileHead.Filename) {
			//重名了
			//改名操作
			var prefix = utils.GetFilePrefix(fileName)
			newPath := fmt.Sprintf("%s_%s.%s", prefix, random.RandStr(4), suffix)
			filePath = path.Join(dirPath, newPath)

		}

		err = os.WriteFile(filePath, imageData, 0666)
		if err != nil {
			response.Response(r, w, nil, err)
			return
		}
		resp.Src = "/" + filePath
		response.Response(r, w, resp, err)
	}
}
