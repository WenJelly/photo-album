package user

import (
	"context"

	commonresponse "photo-album/internal/common/response"
	"photo-album/internal/svc"
	"photo-album/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserVOLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserVOLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserVOLogic {
	return &GetUserVOLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserVOLogic) GetUserVO(req *types.UserIDQueryRequest) (*types.UserProfileResponse, error) {
	if req == nil || req.Id <= 0 {
		return nil, commonresponse.BadRequest("id 必须是正整数")
	}

	return loadActiveUserProfile(l.ctx, l.svcCtx, req.Id.Int64())
}
