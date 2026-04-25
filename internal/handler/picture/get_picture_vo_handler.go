package picture

import (
	"net/http"

	commonrequest "photo-album/internal/common/request"
	"photo-album/internal/common/response"
	logicpicture "photo-album/internal/logic/picture"
	"photo-album/internal/svc"
	"photo-album/internal/types"
)

func GetPictureVOHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetPictureRequest
		if err := commonrequest.ParseJSON(r, &req); err != nil {
			response.Response(w, nil, response.BadRequest(err.Error()))
			return
		}

		l := logicpicture.NewGetPictureVOLogic(r.Context(), svcCtx)
		resp, err := l.GetPictureVO(&req, r.Header.Get("Authorization"))
		response.Response(w, resp, err)
	}
}
