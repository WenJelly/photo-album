package picture

import (
	"net/http"
	"strconv"

	"photo-album/internal/common/response"
	logicpicture "photo-album/internal/logic/picture"
	"photo-album/internal/svc"
	"photo-album/internal/types"
)

func UploadPictureHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(logicpicture.MaxMultipartMemory); err != nil {
			response.Response(w, nil, response.BadRequest("解析上传表单失败"))
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			response.Response(w, nil, response.BadRequest("缺少文件字段 file"))
			return
		}
		defer file.Close()

		req, err := buildUploadRequestFromForm(r)
		if err != nil {
			response.Response(w, nil, err)
			return
		}

		l := logicpicture.NewUploadPictureLogic(r.Context(), svcCtx)
		resp, err := l.UploadPicture(file, header, req, r.Header.Get("Authorization"))
		response.Response(w, resp, err)
	}
}

func buildUploadRequestFromForm(r *http.Request) (*types.PictureUploadRequest, error) {
	var id int64

	rawID := r.FormValue("id")
	if rawID != "" {
		parsedID, err := strconv.ParseInt(rawID, 10, 64)
		if err != nil || parsedID <= 0 {
			return nil, response.BadRequest("id 必须是正整数")
		}
		id = parsedID
	}

	tags, err := logicpicture.ParseTagsInput(r.FormValue("tags"))
	if err != nil {
		return nil, response.BadRequest("tags 必须是 JSON 数组或逗号分隔字符串")
	}

	return &types.PictureUploadRequest{
		Id:           id,
		PicName:      r.FormValue("picName"),
		Introduction: r.FormValue("introduction"),
		Category:     r.FormValue("category"),
		Tags:         tags,
	}, nil
}
