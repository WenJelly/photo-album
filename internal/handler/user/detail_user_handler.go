package user

import (
	"net/http"

	commonrequest "photo-album/internal/common/request"
	"photo-album/internal/common/response"
	logicuser "photo-album/internal/logic/user"
	"photo-album/internal/svc"
	"photo-album/internal/types"
)

func DetailUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DetailUserRequest
		if err := commonrequest.ParseJSON(r, &req); err != nil {
			response.Response(w, nil, response.BadRequest(err.Error()))
			return
		}

		l := logicuser.NewDetailUserLogic(r.Context(), svcCtx)
		resp, err := l.DetailUser(&req, r.Header.Get("Authorization"))
		response.Response(w, resp, err)
	}
}
