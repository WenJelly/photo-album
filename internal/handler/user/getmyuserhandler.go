package user

import (
	"net/http"

	"photo-album/internal/common/response"
	logicuser "photo-album/internal/logic/user"
	"photo-album/internal/svc"
)

func GetMyUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logicuser.NewGetMyUserLogic(r.Context(), svcCtx)
		resp, err := l.GetMyUser(r.Header.Get("Authorization"))
		response.Response(w, resp, err)
	}
}
