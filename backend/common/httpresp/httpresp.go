package httpresp

import (
	"encoding/json"
	"net/http"

	"github.com/Z3Labs/Hackathon/backend/common/errorx"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Http 成功响应
func Http(w http.ResponseWriter, r *http.Request, data interface{}, err error) {
	if err != nil {
		HttpErr(w, r, err)
		return
	}

	resp := Response{
		Code:    200,
		Message: "success",
		Data:    data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// HttpErr 错误响应
func HttpErr(w http.ResponseWriter, r *http.Request, err error) {
	var resp Response

	if statErr, ok := err.(*errorx.StatCodeError); ok {
		resp = Response{
			Code:    statErr.Code,
			Message: statErr.Message,
		}
		w.WriteHeader(statErr.Status)
	} else {
		resp = Response{
			Code:    500,
			Message: err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
