package user

import (
	"net/http"

	commonrequest "photo-album/internal/common/request"
	"photo-album/internal/common/response"
	logicuser "photo-album/internal/logic/user"
	"photo-album/internal/svc"
	"photo-album/internal/types"
)

func UpdateMyUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateMyUserRequest
		if err := commonrequest.ParseJSON(r, &req); err != nil {
			response.Response(w, nil, response.BadRequest(err.Error()))
			return
		}

		l := logicuser.NewUpdateMyUserLogic(r.Context(), svcCtx)
		resp, err := l.UpdateMyUser(&req, r.Header.Get("Authorization"))
		response.Response(w, resp, err)
	}
}
