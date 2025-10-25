package apps

import (
	"net/http"

	"github.com/Z3Labs/Hackathon/backend/internal/logic/apps"
	"github.com/Z3Labs/Hackathon/backend/internal/svc"
	"github.com/Z3Labs/Hackathon/backend/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetAppVersionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetAppVersionsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := apps.NewGetAppVersionsLogic(r.Context(), svcCtx)
		resp, err := l.GetAppVersions(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
