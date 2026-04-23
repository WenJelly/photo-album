package user

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	commonauth "photo-album/internal/common/auth"
	commonresponse "photo-album/internal/common/response"
	"photo-album/internal/svc"
	"photo-album/model"

	"golang.org/x/crypto/bcrypt"
)

const (
	maxUserNameLength    = 256
	maxUserAvatarLength  = 1024
	maxUserProfileLength = 512
)

var nowFunc = func() time.Time {
	return time.Now()
}

type normalizedUserPatch struct {
	userName     string
	userEmail    string
	userAvatar   string
	userProfile  string
	userRole     string
	passwordHash string

	hasUserName     bool
	hasUserEmail    bool
	hasPasswordHash bool
	hasUserAvatar   bool
	hasUserProfile  bool
	hasUserRole     bool
}

func formatUserID(userID int64) string {
	if userID <= 0 {
		return ""
	}

	return strconv.FormatInt(userID, 10)
}

func parseRequiredUserID(raw string) (int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, commonresponse.BadRequest("id 必须是正整数")
	}

	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, commonresponse.BadRequest("id 必须是正整数")
	}

	return id, nil
}

func loadRequiredLoginUser(ctx context.Context, svcCtx *svc.ServiceContext, authorization string) (*model.User, error) {
	return commonauth.LoadRequiredLoginUser(ctx, svcCtx, authorization)
}

func cloneUser(userInfo *model.User) *model.User {
	if userInfo == nil {
		return &model.User{}
	}

	cloned := *userInfo
	return &cloned
}

func normalizeUserPatch(userName, userEmail, userPassword, userAvatar, userProfile, userRole string, allowRole bool) (*normalizedUserPatch, error) {
	resp := &normalizedUserPatch{}

	if userName = strings.TrimSpace(userName); userName != "" {
		if utf8.RuneCountInString(userName) > maxUserNameLength {
			return nil, commonresponse.BadRequest("userName 长度不能超过 256")
		}
		resp.userName = userName
		resp.hasUserName = true
	}

	if userEmail = strings.TrimSpace(userEmail); userEmail != "" {
		resp.userEmail = userEmail
		resp.hasUserEmail = true
	}

	if userPassword = strings.TrimSpace(userPassword); userPassword != "" {
		hashPassword, err := bcrypt.GenerateFromPassword([]byte(userPassword), bcrypt.DefaultCost)
		if err != nil {
			return nil, commonresponse.InternalServerError("密码加密失败")
		}
		resp.passwordHash = string(hashPassword)
		resp.hasPasswordHash = true
	}

	if userAvatar = strings.TrimSpace(userAvatar); userAvatar != "" {
		if len(userAvatar) > maxUserAvatarLength {
			return nil, commonresponse.BadRequest("userAvatar 长度不能超过 1024")
		}
		resp.userAvatar = userAvatar
		resp.hasUserAvatar = true
	}

	if userProfile = strings.TrimSpace(userProfile); userProfile != "" {
		if utf8.RuneCountInString(userProfile) > maxUserProfileLength {
			return nil, commonresponse.BadRequest("userProfile 长度不能超过 512")
		}
		resp.userProfile = userProfile
		resp.hasUserProfile = true
	}

	if allowRole {
		if userRole = strings.TrimSpace(userRole); userRole != "" {
			if userRole != "user" && userRole != "admin" {
				return nil, commonresponse.BadRequest("userRole 只能是 user 或 admin")
			}
			resp.userRole = userRole
			resp.hasUserRole = true
		}
	}

	if !resp.hasUserName && !resp.hasUserEmail && !resp.hasPasswordHash && !resp.hasUserAvatar && !resp.hasUserProfile && !resp.hasUserRole {
		return nil, commonresponse.BadRequest("至少更新一个字段")
	}

	return resp, nil
}

func applyUserPatch(userInfo *model.User, patch *normalizedUserPatch) *model.User {
	updated := cloneUser(userInfo)
	if patch == nil {
		return updated
	}

	if patch.hasUserName {
		updated.UserName = patch.userName
	}
	if patch.hasUserEmail {
		updated.UserEmail = patch.userEmail
	}
	if patch.hasPasswordHash {
		updated.UserPassword = patch.passwordHash
	}
	if patch.hasUserAvatar {
		updated.UserAvatar = patch.userAvatar
	}
	if patch.hasUserProfile {
		updated.UserProfile = patch.userProfile
	}
	if patch.hasUserRole {
		updated.UserRole = patch.userRole
	}

	now := nowFunc()
	updated.EditTime = now
	updated.UpdateTime = now
	return updated
}

func ensureUserEmailAvailable(ctx context.Context, svcCtx *svc.ServiceContext, userID int64, userEmail string) error {
	if strings.TrimSpace(userEmail) == "" {
		return nil
	}

	existing, err := svcCtx.UserModel.FindOneByUserEmail(ctx, userEmail)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil
		}
		return commonresponse.InternalServerError("查询用户邮箱失败")
	}

	if existing.Id != userID && existing.IsDelete == 0 {
		return commonresponse.Conflict("邮箱已注册")
	}

	return nil
}
