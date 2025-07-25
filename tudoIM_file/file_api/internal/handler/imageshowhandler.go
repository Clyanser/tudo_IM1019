package handler

import (
	"net/http"
	"os"
	"path"

	"github.com/zeromicro/go-zero/rest/httpx"
	"tudo_IM1019/tudoIM_file/file_api/internal/svc"
	"tudo_IM1019/tudoIM_file/file_api/internal/types"

	"tudo_IM1019/common/response"
)

func ImageShowHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ImageShowRequest
		if err := httpx.Parse(r, &req); err != nil {
			response.Response(r, w, nil, err)
			return
		}
		filePath := path.Join("uploads", req.ImageType, req.ImageName)
		bytedata, err := os.ReadFile(filePath)
		if err != nil {
			response.Response(r, w, nil, err)
			return
		}
		_, err = w.Write(bytedata)
		if err != nil {
			response.Response(r, w, nil, err)
		}

		//l := logic.NewImageShowLogic(r.Context(), svcCtx)
		//err = l.ImageShow(&req)
		//// if err != nil {
		////  httpx.ErrorCtx(r.Context(), w, err)
		////} else {
		////  httpx.Ok(w)
		//// }
		//response.Response(r, w, nil, err)
	}
}
