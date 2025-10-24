package alert

import (
	"net/http"

	"github.com/Z3Labs/Hackathon/backend/internal/logic/alert"
	"github.com/Z3Labs/Hackathon/backend/internal/svc"
	"github.com/Z3Labs/Hackathon/backend/internal/types"
	"github.com/qbox/jarvis/common/errorx"
	"github.com/qbox/jarvis/common/httpresp"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func AlertCallBackHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PostAlertCallbackReq
		if err := httpx.Parse(r, &req); err != nil {
			httpresp.HttpErr(w, r, errorx.NewStatCodeError(http.StatusBadRequest, 2, err.Error()))
			return
		}

		l := alert.NewAlertCallBackLogic(r.Context(), svcCtx)
		err := l.AlertCallBack(&req)

		httpresp.Http(w, r, nil, err)

	}
}
