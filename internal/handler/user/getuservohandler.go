package user

import (
	"net/http"

	"photo-album/internal/common/response"
	logicuser "photo-album/internal/logic/user"
	"photo-album/internal/svc"
)

func GetUserVOHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := parseUserIDQuery(r.URL.Query().Get("id"))
		if err != nil {
			response.Response(w, nil, err)
			return
		}

		l := logicuser.NewGetUserVOLogic(r.Context(), svcCtx)
		resp, err := l.GetUserVO(req)
		response.Response(w, resp, err)
	}
}
