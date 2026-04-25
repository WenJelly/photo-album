package user

import (
	"context"
	"errors"

	commonresponse "photo-album/internal/common/response"
	"photo-album/internal/svc"
	"photo-album/internal/types"
	"photo-album/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminUpdateUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminUpdateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUpdateUserLogic {
	return &AdminUpdateUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminUpdateUserLogic) AdminUpdateUser(req *types.AdminUpdateUserRequest, _ string) (*types.DetailUserResponse, error) {
	if req == nil {
		return nil, commonresponse.BadRequest("请求体不能为空")
	}

	targetUserID, err := parseRequiredUserID(req.Id)
	if err != nil {
		return nil, err
	}

	targetUser, err := l.svcCtx.UserModel.FindOneActive(l.ctx, targetUserID)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, commonresponse.NotFound("用户不存在")
		}
		return nil, commonresponse.InternalServerError("查询用户信息失败")
	}

	patch, err := normalizeUserPatch(req.UserName, req.UserEmail, req.UserPassword, req.UserAvatar, req.UserProfile, req.UserRole, true)
	if err != nil {
		return nil, err
	}
	if patch.hasUserEmail {
		if err := ensureUserEmailAvailable(l.ctx, l.svcCtx, targetUser.Id, patch.userEmail); err != nil {
			return nil, err
		}
	}

	updatedUser := applyUserPatch(targetUser, patch)
	if err := l.svcCtx.UserModel.Update(l.ctx, updatedUser); err != nil {
		return nil, commonresponse.InternalServerError("更新用户信息失败")
	}

	pictureStats, err := loadUserPictureStats(l.ctx, l.svcCtx, updatedUser.Id)
	if err != nil {
		return nil, err
	}

	return buildDetailUserResponse(updatedUser, pictureStats), nil
}
