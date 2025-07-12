package handler

import (
	"net/http"
	"tudo_IM1019/common/response"
	"tudo_IM1019/tudoIM_auth/auth_api/internal/logic"
	"tudo_IM1019/tudoIM_auth/auth_api/internal/svc"
)

func logoutHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewLogoutLogic(r.Context(), svcCtx)

		token := r.Header.Get("token")

		resp, err := l.Logout(token)
		// if err != nil {
		//  httpx.ErrorCtx(r.Context(), w, err)
		//} else {
		//  httpx.OkJsonCtx(r.Context(), w, resp)
		// }
		response.Response(r, w, resp, err)
	}
}
