package picture

import (
	"context"
	"errors"

	commonresponse "photo-album/internal/common/response"
	"photo-album/internal/svc"
	"photo-album/internal/types"
	"photo-album/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPictureLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPictureLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPictureLogic {
	return &GetPictureLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPictureLogic) GetPicture(req *types.PictureGetRequest, authorization string) (*types.PictureResponse, error) {
	return l.GetPictureVOByID(req.Id, authorization)
}

func (l *GetPictureLogic) GetPictureVOByID(id int64, authorization string) (*types.PictureResponse, error) {
	if id <= 0 {
		return nil, commonresponse.BadRequest("id 必须是正整数")
	}

	pictureInfo, err := l.svcCtx.PicturesModel.FindOneActive(l.ctx, id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, commonresponse.NotFound("图片不存在")
		}
		return nil, commonresponse.InternalServerError("查询图片失败")
	}

	loginUser, err := loadOptionalLoginUser(l.ctx, l.svcCtx, authorization)
	if err != nil {
		return nil, err
	}
	if !canViewPictureDetail(pictureInfo.ReviewStatus, pictureInfo.UserId, loginUser) {
		return nil, commonresponse.Forbidden("当前图片暂不可查看")
	}

	if err := l.svcCtx.PicturesModel.IncrementViewCount(l.ctx, pictureInfo.Id); err != nil {
		return nil, commonresponse.InternalServerError("更新浏览次数失败")
	}
	pictureInfo.ViewCount++

	userMap, err := loadUserSummaryMap(l.ctx, l.svcCtx, []int64{pictureInfo.UserId})
	if err != nil {
		return nil, err
	}

	return buildPictureResponseWithUser(pictureInfo, userMap[pictureInfo.UserId]), nil
}

func (l *GetPictureLogic) GetPictureRawByID(id int64, authorization string) (*types.PictureResponse, error) {
	if id <= 0 {
		return nil, commonresponse.BadRequest("id 必须是正整数")
	}

	if _, err := loadRequiredAdmin(l.ctx, l.svcCtx, authorization); err != nil {
		return nil, err
	}

	pictureInfo, err := l.svcCtx.PicturesModel.FindOneActive(l.ctx, id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, commonresponse.NotFound("图片不存在")
		}
		return nil, commonresponse.InternalServerError("查询图片失败")
	}

	return buildPictureResponse(pictureInfo), nil
}
