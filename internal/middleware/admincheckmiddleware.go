package middleware

import (
	"net/http"

	commonauth "photo-album/internal/common/auth"
	commonresponse "photo-album/internal/common/response"
	"photo-album/internal/svc"
)

type AdminCheckMiddleware struct {
	svcCtx *svc.ServiceContext
}

func NewAdminCheckMiddleware(svcCtx *svc.ServiceContext) *AdminCheckMiddleware {
	return &AdminCheckMiddleware{
		svcCtx: svcCtx,
	}
}

func (m *AdminCheckMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		loginUser, err := commonauth.LoadRequiredAdminUser(r.Context(), m.svcCtx, r.Header.Get("Authorization"))
		if err != nil {
			commonresponse.Response(w, nil, err)
			return
		}

		next(w, r.WithContext(commonauth.WithLoginUser(r.Context(), loginUser)))
	}
}
