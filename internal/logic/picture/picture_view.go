package picture

import (
	"context"

	commonresponse "photo-album/internal/common/response"
	"photo-album/internal/svc"
	"photo-album/internal/types"
	"photo-album/model"
)

func buildPictureResponse(pictureInfo *model.Pictures) *types.PictureResponse {
	return buildPictureResponseWithUser(pictureInfo, nil)
}

func buildPictureResponseWithUser(pictureInfo *model.Pictures, userSummary *types.UserSummary) *types.PictureResponse {
	if pictureInfo == nil {
		return nil
	}

	return &types.PictureResponse{
		Id:            types.NewSnowflakeID(pictureInfo.Id),
		Url:           pictureInfo.Url,
		Name:          pictureInfo.Name,
		Introduction:  nullStringValue(pictureInfo.Introduction),
		Category:      nullStringValue(pictureInfo.Category),
		Tags:          parseStoredTags(pictureInfo.Tags),
		PicSize:       nullInt64Value(pictureInfo.PicSize),
		PicWidth:      nullInt64Value(pictureInfo.PicWidth),
		PicHeight:     nullInt64Value(pictureInfo.PicHeight),
		PicScale:      nullFloat64Value(pictureInfo.PicScale),
		PicFormat:     nullStringValue(pictureInfo.PicFormat),
		UserId:        types.NewSnowflakeID(pictureInfo.UserId),
		User:          userSummary,
		CreateTime:    pictureInfo.CreateTime.Format("2006-01-02 15:04:05"),
		EditTime:      pictureInfo.EditTime.Format("2006-01-02 15:04:05"),
		UpdateTime:    pictureInfo.UpdateTime.Format("2006-01-02 15:04:05"),
		ReviewStatus:  pictureInfo.ReviewStatus,
		ReviewMessage: nullStringValue(pictureInfo.ReviewMessage),
		ReviewerId:    types.NewSnowflakeID(nullInt64Value(pictureInfo.ReviewerId)),
		ReviewTime:    nullTimeValue(pictureInfo.ReviewTime),
		ThumbnailUrl:  nullStringValue(pictureInfo.ThumbnailUrl),
		PicColor:      nullStringValue(pictureInfo.PicColor),
		ViewCount:     pictureInfo.ViewCount,
		LikeCount:     pictureInfo.LikeCount,
	}
}

func buildUserSummary(user *model.User) *types.UserSummary {
	if user == nil {
		return nil
	}

	return &types.UserSummary{
		Id:          types.NewSnowflakeID(user.Id),
		UserName:    user.UserName,
		UserAvatar:  user.UserAvatar,
		UserProfile: user.UserProfile,
		UserRole:    user.UserRole,
	}
}

func loadUserSummaryMap(ctx context.Context, svcCtx *svc.ServiceContext, userIDs []int64) (map[int64]*types.UserSummary, error) {
	users, err := svcCtx.UserModel.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, commonresponse.InternalServerError("查询创建人信息失败")
	}

	userMap := make(map[int64]*types.UserSummary, len(users))
	for _, user := range users {
		if user == nil {
			continue
		}
		userMap[user.Id] = buildUserSummary(user)
	}

	return userMap, nil
}

func collectPictureUserIDs(pictures []*model.Pictures) []int64 {
	if len(pictures) == 0 {
		return nil
	}

	seen := make(map[int64]struct{}, len(pictures))
	userIDs := make([]int64, 0, len(pictures))
	for _, pictureInfo := range pictures {
		if pictureInfo == nil || pictureInfo.UserId <= 0 {
			continue
		}
		if _, ok := seen[pictureInfo.UserId]; ok {
			continue
		}
		seen[pictureInfo.UserId] = struct{}{}
		userIDs = append(userIDs, pictureInfo.UserId)
	}

	return userIDs
}

func buildPictureListResponse(ctx context.Context, svcCtx *svc.ServiceContext, pictures []*model.Pictures, withUser bool) ([]*types.PictureResponse, error) {
	if len(pictures) == 0 {
		return []*types.PictureResponse{}, nil
	}

	var userMap map[int64]*types.UserSummary
	var err error
	if withUser {
		userMap, err = loadUserSummaryMap(ctx, svcCtx, collectPictureUserIDs(pictures))
		if err != nil {
			return nil, err
		}
	}

	resp := make([]*types.PictureResponse, 0, len(pictures))
	for _, pictureInfo := range pictures {
		if pictureInfo == nil {
			continue
		}
		resp = append(resp, buildPictureResponseWithUser(pictureInfo, userMap[pictureInfo.UserId]))
	}

	return resp, nil
}
