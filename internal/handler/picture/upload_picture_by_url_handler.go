package picture

import (
	"net/http"

	commonrequest "photo-album/internal/common/request"
	"photo-album/internal/common/response"
	logicpicture "photo-album/internal/logic/picture"
	"photo-album/internal/svc"
	"photo-album/internal/types"
)

func UploadPictureByUrlHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PictureUploadByUrlRequest
		if err := commonrequest.ParseJSON(r, &req); err != nil {
			response.Response(w, nil, response.BadRequest(err.Error()))
			return
		}

		l := logicpicture.NewUploadPictureByUrlLogic(r.Context(), svcCtx)
		resp, err := l.UploadPictureByUrl(&req, r.Header.Get("Authorization"))
		response.Response(w, resp, err)
	}
}
