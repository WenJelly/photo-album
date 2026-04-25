// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"net/http"
	"photo-album/internal/common/response"
	"photo-album/internal/logic/user"
	"photo-album/internal/svc"
	"photo-album/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func LoginUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LoginRequest
		if err := httpx.Parse(r, &req); err != nil {
			response.Response(w, nil, response.BadRequest(err.Error()))
			return
		}

		l := user.NewLoginUserLogic(r.Context(), svcCtx)
		resp, err := l.LoginUser(&req)
		response.Response(w, resp, err)
	}
}
