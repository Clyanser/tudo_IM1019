package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"tudo_IM1019/tudoIM_user/user_api/internal/logic"
	"tudo_IM1019/tudoIM_user/user_api/internal/svc"
	"tudo_IM1019/tudoIM_user/user_api/internal/types"

	"tudo_IM1019/common/response"
)

func UserVerifyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UserVerifyRequest
		if err := httpx.Parse(r, &req); err != nil {
			response.Response(r, w, nil, err)
			return
		}

		l := logic.NewUserVerifyLogic(r.Context(), svcCtx)
		resp, err := l.UserVerify(&req)
		// if err != nil {
		//  httpx.ErrorCtx(r.Context(), w, err)
		//} else {
		//  httpx.OkJsonCtx(r.Context(), w, resp)
		// }
		response.Response(r, w, resp, err)
	}
}
