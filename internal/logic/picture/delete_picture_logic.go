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

func (l *DeletePictureLogic) DeletePicture(req *types.DeletePictureRequest, authorization string) error {
	if req == nil {
		return commonresponse.BadRequest("请求体不能为空")
	}

	pictureID, err := parseRequiredSnowflakeID(req.Id, "id")
	if err != nil {
		return err
	}

	loginUser, err := loadRequiredLoginUser(l.ctx, l.svcCtx, authorization)
	if err != nil {
		return err
	}

	pictureInfo, err := l.svcCtx.PicturesModel.FindOneActive(l.ctx, pictureID)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return commonresponse.NotFound("图片不存在")
		}
		return commonresponse.InternalServerError("查询图片失败")
	}

	if !canManagePicture(pictureInfo.UserId, loginUser) {
		return commonresponse.Forbidden("无权删除该图片")
	}

	if objectKey, ok := extractObjectKeyFromURL(l.svcCtx.Config.Cos.Host, pictureInfo.Url); ok {
		if err := deleteFileFromCOS(l.ctx, l.svcCtx, objectKey); err != nil {
			return err
		}
	}

	pictureInfo.IsDelete = 1
	pictureInfo.UpdateTime = time.Now()

	if err := l.svcCtx.PicturesModel.Update(l.ctx, pictureInfo); err != nil {
		return commonresponse.InternalServerError("删除图片失败")
	}

	return nil
}
