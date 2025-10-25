package machines

import (
	"net/http"

	"github.com/Z3Labs/Hackathon/backend/common/errorx"
	"github.com/Z3Labs/Hackathon/backend/common/httpresp"
	"github.com/Z3Labs/Hackathon/backend/internal/logic/machines"
	"github.com/Z3Labs/Hackathon/backend/internal/svc"
	"github.com/Z3Labs/Hackathon/backend/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func DeleteMachineHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DeleteMachineReq
		if err := httpx.Parse(r, &req); err != nil {
			httpresp.HttpErr(w, r, errorx.NewStatCodeError(http.StatusBadRequest, 2, err.Error()))
			return
		}

		l := machines.NewDeleteMachineLogic(r.Context(), svcCtx)
		resp, err := l.DeleteMachine(&req)

		httpresp.Http(w, r, resp, err)

	}
}
