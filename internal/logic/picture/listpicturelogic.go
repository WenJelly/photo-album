package picture

import (
	"context"

	commonresponse "photo-album/internal/common/response"
	"photo-album/internal/svc"
	"photo-album/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListPictureLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListPictureLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPictureLogic {
	return &ListPictureLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListPictureLogic) ListPictureVO(req *types.PictureListRequest) (*types.PicturePageResponse, error) {
	return l.listPictures(req, "", true, true)
}

func (l *ListPictureLogic) ListMyPictures(req *types.PictureListRequest, authorization string) (*types.PicturePageResponse, error) {
	loginUser, err := loadRequiredLoginUser(l.ctx, l.svcCtx, authorization)
	if err != nil {
		return nil, err
	}

	if req == nil {
		req = &types.PictureListRequest{}
	} else {
		cloned := *req
		req = &cloned
	}
	req.UserId = types.NewSnowflakeID(loginUser.Id)

	return l.listPictures(req, "", false, true)
}

func (l *ListPictureLogic) ListPictureRaw(req *types.PictureListRequest, authorization string) (*types.PicturePageResponse, error) {
	if _, err := loadRequiredAdmin(l.ctx, l.svcCtx, authorization); err != nil {
		return nil, err
	}

	return l.listPictures(req, authorization, false, false)
}

func (l *ListPictureLogic) listPictures(req *types.PictureListRequest, _ string, publicOnly bool, withUser bool) (*types.PicturePageResponse, error) {
	if req == nil {
		req = &types.PictureListRequest{}
	}

	pageNum, pageSize, err := normalizePicturePage(req.PageNum, req.PageSize)
	if err != nil {
		return nil, err
	}

	whereSQL, args, err := buildPictureListWhere(req, publicOnly)
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
			List:     []*types.PictureResponse{},
		}, nil
	}

	pictures, err := l.svcCtx.PicturesModel.FindByWhere(l.ctx, whereSQL, "`id` desc", pageSize, (pageNum-1)*pageSize, args...)
	if err != nil {
		return nil, commonresponse.InternalServerError("分页查询图片失败")
	}

	list, err := buildPictureListResponse(l.ctx, l.svcCtx, pictures, withUser)
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
