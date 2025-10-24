package deployments

import (
	"net/http"

	"github.com/Z3Labs/Hackathon/backend/common/errorx"
	"github.com/Z3Labs/Hackathon/backend/common/httpresp"
	"github.com/Z3Labs/Hackathon/backend/internal/logic/deployments"
	"github.com/Z3Labs/Hackathon/backend/internal/svc"
	"github.com/Z3Labs/Hackathon/backend/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetDeploymentDetailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetDeploymentDetailReq
		if err := httpx.Parse(r, &req); err != nil {
			httpresp.HttpErr(w, r, errorx.NewStatCodeError(http.StatusBadRequest, 2, err.Error()))
			return
		}

		l := deployments.NewGetDeploymentDetailLogic(r.Context(), svcCtx)
		resp, err := l.GetDeploymentDetail(&req)

		httpresp.Http(w, r, resp, err)

	}
}
