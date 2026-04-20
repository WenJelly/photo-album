package picture

import (
	"net/http"

	"photo-album/internal/common/response"
	logicpicture "photo-album/internal/logic/picture"
	"photo-album/internal/svc"
)

func GetPictureAdminHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := parsePictureIDQuery(r.URL.Query().Get("id"))
		if err != nil {
			response.Response(w, nil, err)
			return
		}

		l := logicpicture.NewGetPictureLogic(r.Context(), svcCtx)
		resp, err := l.GetPictureRawByID(req.Id.Int64(), r.Header.Get("Authorization"))
		response.Response(w, resp, err)
	}
}
