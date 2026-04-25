package picture

import (
	"net/http"

	commonrequest "photo-album/internal/common/request"
	"photo-album/internal/common/response"
	"photo-album/internal/logic/picture"
	"photo-album/internal/svc"
	"photo-album/internal/types"
)

func GetAdminPictureListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminQueryPictureRequest
		if err := commonrequest.ParseJSON(r, &req); err != nil {
			response.Response(w, nil, response.BadRequest(err.Error()))
			return
		}

		l := picture.NewGetAdminPictureListLogic(r.Context(), svcCtx)
		resp, err := l.GetAdminPictureList(&req)
		response.Response(w, resp, err)
	}
}
