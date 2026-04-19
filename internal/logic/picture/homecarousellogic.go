package picture

import (
	"context"

	commonresponse "photo-album/internal/common/response"
	"photo-album/internal/svc"
	"photo-album/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type HomeCarouselLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHomeCarouselLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HomeCarouselLogic {
	return &HomeCarouselLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HomeCarouselLogic) GetHomeCarousel() (*types.PictureCarouselResponse, error) {
	whereSQL, args := buildPublicPictureListWhere("", nil)

	pictures, err := l.svcCtx.PicturesModel.FindByWhere(l.ctx, whereSQL, "`viewCount` desc, `id` desc", 6, 0, args...)
	if err != nil {
		return nil, commonresponse.InternalServerError("查询首页轮播图失败")
	}

	list, err := buildPictureListResponse(l.ctx, l.svcCtx, pictures, true)
	if err != nil {
		return nil, err
	}

	return &types.PictureCarouselResponse{List: list}, nil
}
