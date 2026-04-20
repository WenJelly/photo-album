package picture

import (
	"context"
	"errors"
	"time"

	commonresponse "photo-album/internal/common/response"
	"photo-album/internal/svc"
	"photo-album/internal/types"
	"photo-album/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeletePictureLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeletePictureLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeletePictureLogic {
	return &DeletePictureLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeletePictureLogic) DeletePicture(req *types.PictureDeleteRequest, authorization string) (*types.PictureDeleteResponse, error) {
	if req == nil || req.Id <= 0 {
		return nil, commonresponse.BadRequest("id 必须是正整数")
	}

	loginUser, err := loadRequiredLoginUser(l.ctx, l.svcCtx, authorization)
	if err != nil {
		return nil, err
	}

	pictureInfo, err := l.svcCtx.PicturesModel.FindOneActive(l.ctx, req.Id.Int64())
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, commonresponse.NotFound("图片不存在")
		}
		return nil, commonresponse.InternalServerError("查询图片失败")
	}

	if !canManagePicture(pictureInfo.UserId, loginUser) {
		return nil, commonresponse.Forbidden("无权删除该图片")
	}

	if objectKey, ok := extractObjectKeyFromURL(l.svcCtx.Config.Cos.Host, pictureInfo.Url); ok {
		if err := deleteFileFromCOS(l.ctx, l.svcCtx, objectKey); err != nil {
			return nil, err
		}
	}

	pictureInfo.IsDelete = 1
	pictureInfo.UpdateTime = time.Now()

	if err := l.svcCtx.PicturesModel.Update(l.ctx, pictureInfo); err != nil {
		return nil, commonresponse.InternalServerError("删除图片失败")
	}

	return &types.PictureDeleteResponse{Id: types.NewSnowflakeID(pictureInfo.Id)}, nil
}
