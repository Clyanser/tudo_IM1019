package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"tudo_IM1019/tudoIM_chat/chat_api/internal/logic"
	"tudo_IM1019/tudoIM_chat/chat_api/internal/svc"
	"tudo_IM1019/tudoIM_chat/chat_api/internal/types"

	"tudo_IM1019/common/response"
)

func chatSessionHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ChatSessionRequest
		if err := httpx.Parse(r, &req); err != nil {
			response.Response(r, w, nil, err)
			return
		}

		l := logic.NewChatSessionLogic(r.Context(), svcCtx)
		resp, err := l.ChatSession(&req)
		// if err != nil {
		//  httpx.ErrorCtx(r.Context(), w, err)
		//} else {
		//  httpx.OkJsonCtx(r.Context(), w, resp)
		// }
		response.Response(r, w, resp, err)
	}
}
