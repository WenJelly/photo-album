package user

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	commonauth "photo-album/internal/common/auth"
	commonresponse "photo-album/internal/common/response"
	"photo-album/internal/svc"
	"photo-album/internal/types"
	"photo-album/model"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	maxUserNameLength    = 256
	maxUserAvatarLength  = 1024
	maxUserProfileLength = 512
)

type UpdateMyUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

type normalizedUpdateMyUserRequest struct {
	userName       string
	userAvatar     string
	userProfile    string
	hasUserName    bool
	hasUserAvatar  bool
	hasUserProfile bool
}

func NewUpdateMyUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateMyUserLogic {
	return &UpdateMyUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateMyUserLogic) UpdateMyUser(req *types.UpdateMyUserRequest, authorization string) (*types.UserProfileResponse, error) {
	loginUser, err := commonauth.LoadRequiredLoginUser(l.ctx, l.svcCtx, authorization)
	if err != nil {
		return nil, err
	}

	normalizedReq, err := normalizeUpdateMyUserRequest(req)
	if err != nil {
		return nil, err
	}

	updatedUser := cloneUser(loginUser)
	if normalizedReq.hasUserName {
		updatedUser.UserName = normalizedReq.userName
	}
	if normalizedReq.hasUserAvatar {
		updatedUser.UserAvatar = normalizedReq.userAvatar
	}
	if normalizedReq.hasUserProfile {
		updatedUser.UserProfile = normalizedReq.userProfile
	}

	now := nowFunc()
	updatedUser.EditTime = now
	updatedUser.UpdateTime = now

	if err := l.svcCtx.UserModel.Update(l.ctx, updatedUser); err != nil {
		return nil, commonresponse.InternalServerError("更新用户信息失败")
	}

	pictureStats, err := loadUserPictureStats(l.ctx, l.svcCtx, updatedUser.Id)
	if err != nil {
		return nil, err
	}

	return buildUserProfileResponse(updatedUser, pictureStats), nil
}

var nowFunc = func() time.Time {
	return time.Now()
}

func cloneUser(userInfo *model.User) *model.User {
	if userInfo == nil {
		return &model.User{}
	}

	cloned := *userInfo
	return &cloned
}

func normalizeUpdateMyUserRequest(req *types.UpdateMyUserRequest) (*normalizedUpdateMyUserRequest, error) {
	if req == nil {
		return nil, commonresponse.BadRequest("请求体不能为空")
	}

	resp := &normalizedUpdateMyUserRequest{}

	if req.UserName != nil {
		resp.hasUserName = true
		resp.userName = strings.TrimSpace(*req.UserName)
		if resp.userName == "" {
			return nil, commonresponse.BadRequest("userName 不能为空")
		}
		if utf8.RuneCountInString(resp.userName) > maxUserNameLength {
			return nil, commonresponse.BadRequest("userName 长度不能超过 256")
		}
	}

	if req.UserAvatar != nil {
		resp.hasUserAvatar = true
		resp.userAvatar = strings.TrimSpace(*req.UserAvatar)
		if len(resp.userAvatar) > maxUserAvatarLength {
			return nil, commonresponse.BadRequest("userAvatar 长度不能超过 1024")
		}
	}

	if req.UserProfile != nil {
		resp.hasUserProfile = true
		resp.userProfile = strings.TrimSpace(*req.UserProfile)
		if utf8.RuneCountInString(resp.userProfile) > maxUserProfileLength {
			return nil, commonresponse.BadRequest("userProfile 长度不能超过 512")
		}
	}

	if !resp.hasUserName && !resp.hasUserAvatar && !resp.hasUserProfile {
		return nil, commonresponse.BadRequest("至少更新一个字段")
	}

	return resp, nil
}
