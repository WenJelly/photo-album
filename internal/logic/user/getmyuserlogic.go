package user

import (
	"context"

	commonauth "photo-album/internal/common/auth"
	"photo-album/internal/svc"
	"photo-album/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMyUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMyUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyUserLogic {
	return &GetMyUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMyUserLogic) GetMyUser(authorization string) (*types.UserProfileResponse, error) {
	loginUser, err := commonauth.LoadRequiredLoginUser(l.ctx, l.svcCtx, authorization)
	if err != nil {
		return nil, err
	}

	pictureStats, err := loadUserPictureStats(l.ctx, l.svcCtx, loginUser.Id)
	if err != nil {
		return nil, err
	}

	return buildUserProfileResponse(loginUser, pictureStats), nil
}
