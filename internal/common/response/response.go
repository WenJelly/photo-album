package response

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type Body struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func Response(w http.ResponseWriter, resp interface{}, err error) {
	if err != nil {
		statusCode, body := ErrorBody(err)
		httpx.WriteJson(w, statusCode, body)
		return
	}

	httpx.WriteJson(w, SuccessCode.StatusCode(), Body{
		Code:    SuccessCode.Code(),
		Message: SuccessCode.Message(),
		Data:    resp,
	})
}
