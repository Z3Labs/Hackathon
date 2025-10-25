// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package deployments

import (
	"net/http"

	"github.com/Z3Labs/Hackathon/backend/internal/logic/deployments"
	"github.com/Z3Labs/Hackathon/backend/internal/svc"
	"github.com/Z3Labs/Hackathon/backend/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 取消发布中的设备
func CancelNodeDeploymentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CancelNodeDeploymentReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := deployments.NewCancelNodeDeploymentLogic(r.Context(), svcCtx)
		resp, err := l.CancelNodeDeployment(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
