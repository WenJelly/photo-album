package picture

import (
	"context"

	commonresponse "photo-album/internal/common/response"
	"photo-album/internal/svc"
	"photo-album/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAdminPictureListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAdminPictureListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAdminPictureListLogic {
	return &GetAdminPictureListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAdminPictureListLogic) GetAdminPictureList(req *types.AdminQueryPictureRequest) (*types.PicturePageResponse, error) {
	if req == nil {
		req = &types.AdminQueryPictureRequest{}
	}

	pageNum, pageSize, err := normalizePicturePage(req.PageNum, req.PageSize)
	if err != nil {
		return nil, err
	}

	whereSQL, args, err := buildAdminPictureListWhere(req)
	if err != nil {
		return nil, err
	}

	total, err := l.svcCtx.PicturesModel.CountByWhere(l.ctx, whereSQL, args...)
	if err != nil {
		return nil, commonresponse.InternalServerError("查询图片总数失败")
	}

	if total == 0 {
		return &types.PicturePageResponse{
			PageNum:  pageNum,
			PageSize: pageSize,
			Total:    0,
			List:     []types.PictureResponse{},
		}, nil
	}

	pictures, err := l.svcCtx.PicturesModel.FindByWhere(l.ctx, whereSQL, "`id` desc", pageSize, (pageNum-1)*pageSize, args...)
	if err != nil {
		return nil, commonresponse.InternalServerError("分页查询图片失败")
	}

	list, err := buildPictureListResponse(l.ctx, l.svcCtx, pictures, true, req.CompressPictureType)
	if err != nil {
		return nil, err
	}

	return &types.PicturePageResponse{
		PageNum:  pageNum,
		PageSize: pageSize,
		Total:    total,
		List:     list,
	}, nil
}
