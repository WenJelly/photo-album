package user

import (
	"net/http"

	"photo-album/internal/common/response"
	logicuser "photo-album/internal/logic/user"
	"photo-album/internal/svc"
)

func UploadUserAvatarHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(logicuser.MaxAvatarMultipartMemory); err != nil {
			response.Response(w, nil, response.BadRequest("解析上传表单失败"))
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			response.Response(w, nil, response.BadRequest("缺少文件字段 file"))
			return
		}
		defer file.Close()

		l := logicuser.NewUploadUserAvatarLogic(r.Context(), svcCtx)
		resp, err := l.UploadUserAvatar(file, header, r.Header.Get("Authorization"))
		response.Response(w, resp, err)
	}
}
