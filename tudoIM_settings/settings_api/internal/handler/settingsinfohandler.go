package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"tudo_IM1019/tudoIM_settings/settings_api/internal/logic"
	"tudo_IM1019/tudoIM_settings/settings_api/internal/svc"
	"tudo_IM1019/tudoIM_settings/settings_api/internal/types"

	"tudo_IM1019/common/response"
)

func settingsInfoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SettingsInfoRequest
		if err := httpx.Parse(r, &req); err != nil {
			response.Response(r, w, nil, err)
			return
		}

		l := logic.NewSettingsInfoLogic(r.Context(), svcCtx)
		resp, err := l.SettingsInfo(&req)
		// if err != nil {
		//  httpx.ErrorCtx(r.Context(), w, err)
		//} else {
		//  httpx.OkJsonCtx(r.Context(), w, resp)
		// }
		response.Response(r, w, resp, err)
	}
}
