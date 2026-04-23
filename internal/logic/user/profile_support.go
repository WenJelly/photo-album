package user

import (
	"context"
	"errors"

	commonresponse "photo-album/internal/common/response"
	"photo-album/internal/svc"
	"photo-album/internal/types"
	"photo-album/model"
)

func buildDetailUserResponse(userInfo *model.User, pictureStats *model.PictureStats) *types.DetailUserResponse {
	if userInfo == nil {
		return nil
	}
	if pictureStats == nil {
		pictureStats = &model.PictureStats{}
	}

	return &types.DetailUserResponse{
		Id:                   formatUserID(userInfo.Id),
		UserName:             userInfo.UserName,
		UserEmail:            userInfo.UserEmail,
		UserAvatar:           userInfo.UserAvatar,
		UserProfile:          userInfo.UserProfile,
		UserRole:             userInfo.UserRole,
		CreateTime:           userInfo.CreateTime.Format("2006-01-02 15:04:05"),
		UpdateTime:           userInfo.UpdateTime.Format("2006-01-02 15:04:05"),
		PictureCount:         pictureStats.Total,
		ApprovedPictureCount: pictureStats.ApprovedCount,
		PendingPictureCount:  pictureStats.PendingCount,
		RejectedPictureCount: pictureStats.RejectedCount,
	}
}

func loadUserPictureStats(ctx context.Context, svcCtx *svc.ServiceContext, userID int64) (*model.PictureStats, error) {
	pictureStats, err := svcCtx.PicturesModel.CountStatsByUser(ctx, userID)
	if err != nil {
		return nil, commonresponse.InternalServerError("查询用户作品统计失败")
	}

	return pictureStats, nil
}

func loadActiveUserDetail(ctx context.Context, svcCtx *svc.ServiceContext, userID int64) (*types.DetailUserResponse, error) {
	userInfo, err := svcCtx.UserModel.FindOneActive(ctx, userID)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, commonresponse.NotFound("用户不存在")
		}
		return nil, commonresponse.InternalServerError("查询用户信息失败")
	}

	pictureStats, err := loadUserPictureStats(ctx, svcCtx, userID)
	if err != nil {
		return nil, err
	}

	return buildDetailUserResponse(userInfo, pictureStats), nil
}
