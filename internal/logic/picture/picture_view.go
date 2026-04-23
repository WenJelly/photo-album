package picture

import (
	"context"
	"strconv"

	commonresponse "photo-album/internal/common/response"
	"photo-album/internal/svc"
	"photo-album/internal/types"
	"photo-album/model"
)

func buildPictureResponseWithUser(pictureInfo *model.Pictures, userDetail types.UserDetail, compressOption types.CompressPictureType) (*types.PictureResponse, error) {
	if pictureInfo == nil {
		return nil, nil
	}

	thumbnailURL, err := buildPictureThumbnailURL(pictureInfo.Url, nullInt64Value(pictureInfo.PicSize), compressOption)
	if err != nil {
		return nil, err
	}

	return &types.PictureResponse{
		Id:            formatSnowflakeID(pictureInfo.Id),
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
		UserId:        formatSnowflakeID(pictureInfo.UserId),
		User:          userDetail,
		CreateTime:    pictureInfo.CreateTime.Format("2006-01-02 15:04:05"),
		EditTime:      pictureInfo.EditTime.Format("2006-01-02 15:04:05"),
		UpdateTime:    pictureInfo.UpdateTime.Format("2006-01-02 15:04:05"),
		ReviewStatus:  pictureInfo.ReviewStatus,
		ReviewMessage: nullStringValue(pictureInfo.ReviewMessage),
		ReviewerId:    formatSnowflakeID(nullInt64Value(pictureInfo.ReviewerId)),
		ReviewTime:    nullTimeValue(pictureInfo.ReviewTime),
		ThumbnailUrl:  thumbnailURL,
		PicColor:      nullStringValue(pictureInfo.PicColor),
		ViewCount:     pictureInfo.ViewCount,
		LikeCount:     pictureInfo.LikeCount,
	}, nil
}

func buildUserDetail(user *model.User) types.UserDetail {
	if user == nil {
		return types.UserDetail{}
	}

	return types.UserDetail{
		Id:          formatSnowflakeID(user.Id),
		UserName:    user.UserName,
		UserAvatar:  user.UserAvatar,
		UserProfile: user.UserProfile,
		UserRole:    user.UserRole,
	}
}

func loadUserDetailMap(ctx context.Context, svcCtx *svc.ServiceContext, userIDs []int64) (map[int64]types.UserDetail, error) {
	users, err := svcCtx.UserModel.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, commonresponse.InternalServerError("查询创建人信息失败")
	}

	userMap := make(map[int64]types.UserDetail, len(users))
	for _, user := range users {
		if user == nil {
			continue
		}
		userMap[user.Id] = buildUserDetail(user)
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

func buildPictureListResponse(ctx context.Context, svcCtx *svc.ServiceContext, pictures []*model.Pictures, withUser bool, compressOption types.CompressPictureType) ([]types.PictureResponse, error) {
	if len(pictures) == 0 {
		return []types.PictureResponse{}, nil
	}

	userMap := map[int64]types.UserDetail{}
	var err error
	if withUser {
		userMap, err = loadUserDetailMap(ctx, svcCtx, collectPictureUserIDs(pictures))
		if err != nil {
			return nil, err
		}
	}

	resp := make([]types.PictureResponse, 0, len(pictures))
	for _, pictureInfo := range pictures {
		if pictureInfo == nil {
			continue
		}

		item, err := buildPictureResponseWithUser(pictureInfo, userMap[pictureInfo.UserId], compressOption)
		if err != nil {
			return nil, err
		}
		resp = append(resp, *item)
	}

	return resp, nil
}

func formatSnowflakeID(value int64) string {
	if value <= 0 {
		return ""
	}

	return strconv.FormatInt(value, 10)
}
