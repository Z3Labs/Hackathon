package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"hackathon/backend/internal/logic"
	"hackathon/backend/internal/svc"
	"hackathon/backend/internal/types"
)

func PingHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PingReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewPingLogic(r.Context(), svcCtx)
		resp, err := l.Ping(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
