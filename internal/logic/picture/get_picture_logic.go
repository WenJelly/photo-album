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

type GetPictureVOLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPictureVOLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPictureVOLogic {
	return &GetPictureVOLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPictureVOLogic) GetPictureVO(req *types.GetPictureRequest, authorization string) (*types.PictureResponse, error) {
	if req == nil {
		return nil, commonresponse.BadRequest("请求体不能为空")
	}

	id, err := parseRequiredSnowflakeID(req.Id, "id")
	if err != nil {
		return nil, err
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

	userMap, err := loadUserDetailMap(l.ctx, l.svcCtx, []int64{pictureInfo.UserId})
	if err != nil {
		return nil, err
	}

	return buildPictureResponseWithUser(pictureInfo, userMap[pictureInfo.UserId], req.CompressPictureType)
}
