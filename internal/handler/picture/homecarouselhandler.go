package picture

import (
	"net/http"

	"photo-album/internal/common/response"
	logicpicture "photo-album/internal/logic/picture"
	"photo-album/internal/svc"
)

func HomeCarouselHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logicpicture.NewHomeCarouselLogic(r.Context(), svcCtx)
		resp, err := l.GetHomeCarousel()
		response.Response(w, resp, err)
	}
}
