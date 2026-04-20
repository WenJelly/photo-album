package picture

import (
	"context"
	"mime/multipart"

	"photo-album/internal/svc"
	"photo-album/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UploadPictureLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUploadPictureLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadPictureLogic {
	return &UploadPictureLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadPictureLogic) UploadPicture(file multipart.File, header *multipart.FileHeader, req *types.PictureUploadRequest, authorization string) (*types.PictureResponse, error) {
	loginUser, err := loadRequiredLoginUser(l.ctx, l.svcCtx, authorization)
	if err != nil {
		return nil, err
	}

	tempPath, originalFilename, cleanup, err := saveMultipartFileToTemp(file, header)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	return storePicture(l.ctx, l.svcCtx, tempPath, originalFilename, pictureWriteRequest{
		ID:           req.Id.Int64(),
		PicName:      req.PicName,
		Introduction: req.Introduction,
		Category:     req.Category,
		Tags:         req.Tags,
	}, loginUser)
}
