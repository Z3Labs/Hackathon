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

// 发布指定设备
func DeployNodeDeploymentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DeployNodeDeploymentReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := deployments.NewDeployNodeDeploymentLogic(r.Context(), svcCtx)
		resp, err := l.DeployNodeDeployment(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
