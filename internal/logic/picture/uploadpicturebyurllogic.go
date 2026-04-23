package picture

import (
	"context"
	"strings"

	"photo-album/internal/common/response"
	"photo-album/internal/svc"
	"photo-album/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UploadPictureByUrlLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUploadPictureByUrlLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadPictureByUrlLogic {
	return &UploadPictureByUrlLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadPictureByUrlLogic) UploadPictureByUrl(req *types.PictureUploadByUrlRequest, authorization string) (*types.PictureResponse, error) {
	if req == nil {
		return nil, response.BadRequest("请求体不能为空")
	}
	if strings.TrimSpace(req.FileUrl) == "" {
		return nil, response.BadRequest("fileUrl 不能为空")
	}

	pictureID, err := parseOptionalSnowflakeID(req.Id, "id")
	if err != nil {
		return nil, err
	}

	loginUser, err := loadRequiredLoginUser(l.ctx, l.svcCtx, authorization)
	if err != nil {
		return nil, err
	}

	tempPath, originalFilename, cleanup, err := downloadRemoteImageToTemp(l.ctx, req.FileUrl)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	return storePicture(l.ctx, l.svcCtx, tempPath, originalFilename, pictureWriteRequest{
		ID:           pictureID,
		PicName:      req.PicName,
		Introduction: req.Introduction,
		Category:     req.Category,
		Tags:         req.Tags,
	}, loginUser)
}
