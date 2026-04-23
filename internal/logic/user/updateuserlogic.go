package user

import (
	"context"

	commonresponse "photo-album/internal/common/response"
	"photo-album/internal/svc"
	"photo-album/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserLogic {
	return &UpdateUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateUserLogic) UpdateUser(req *types.UpdateUserRequest, authorization string) (*types.DetailUserResponse, error) {
	if req == nil {
		return nil, commonresponse.BadRequest("请求体不能为空")
	}

	loginUser, err := loadRequiredLoginUser(l.ctx, l.svcCtx, authorization)
	if err != nil {
		return nil, err
	}

	targetUserID, err := parseRequiredUserID(req.Id)
	if err != nil {
		return nil, err
	}
	if targetUserID != loginUser.Id {
		return nil, commonresponse.Forbidden("只能修改自己的信息")
	}

	patch, err := normalizeUserPatch(req.UserName, req.UserEmail, req.UserPassword, req.UserAvatar, req.UserProfile, "", false)
	if err != nil {
		return nil, err
	}
	if patch.hasUserEmail {
		if err := ensureUserEmailAvailable(l.ctx, l.svcCtx, loginUser.Id, patch.userEmail); err != nil {
			return nil, err
		}
	}

	updatedUser := applyUserPatch(loginUser, patch)
	if err := l.svcCtx.UserModel.Update(l.ctx, updatedUser); err != nil {
		return nil, commonresponse.InternalServerError("更新用户信息失败")
	}

	pictureStats, err := loadUserPictureStats(l.ctx, l.svcCtx, updatedUser.Id)
	if err != nil {
		return nil, err
	}

	return buildDetailUserResponse(updatedUser, pictureStats), nil
}
