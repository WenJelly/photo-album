package user

import (
	"context"
	"database/sql"
	"testing"
	"time"

	commonresponse "photo-album/internal/common/response"
	"photo-album/internal/config"
	"photo-album/internal/svc"
	"photo-album/internal/types"
	"photo-album/model"

	"github.com/golang-jwt/jwt/v4"
)

func TestGetUserVOReturnsProfileWithStats(t *testing.T) {
	userModel := &stubUserModel{
		activeUser: &model.User{
			Id:         7,
			UserName:   "测试用户",
			UserAvatar: "https://example.com/avatar.png",
			UserRole:   "user",
			CreateTime: time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
			UpdateTime: time.Date(2026, 4, 20, 11, 0, 0, 0, time.UTC),
		},
	}
	picturesModel := &stubPicturesModel{
		stats: &model.PictureStats{
			Total:         8,
			ApprovedCount: 5,
			PendingCount:  2,
			RejectedCount: 1,
		},
	}

	logic := NewGetUserVOLogic(context.Background(), newUserTestServiceContext("", userModel, picturesModel))
	resp, err := logic.GetUserVO(&types.UserIDQueryRequest{Id: types.NewSnowflakeID(7)})
	if err != nil {
		t.Fatalf("GetUserVO() unexpected error = %v", err)
	}
	if resp == nil {
		t.Fatalf("GetUserVO() returned nil response")
	}
	if resp.Id.Int64() != 7 {
		t.Fatalf("GetUserVO() id = %d, want 7", resp.Id)
	}
	if resp.PictureCount != 8 || resp.ApprovedPictureCount != 5 || resp.PendingPictureCount != 2 || resp.RejectedPictureCount != 1 {
		t.Fatalf("GetUserVO() stats = %+v, want total=8 approved=5 pending=2 rejected=1", resp)
	}
}

func TestGetMyUserRequiresLogin(t *testing.T) {
	logic := NewGetMyUserLogic(context.Background(), newUserTestServiceContext("secret", &stubUserModel{}, &stubPicturesModel{}))

	_, err := logic.GetMyUser("")
	if err == nil {
		t.Fatalf("GetMyUser() expected error")
	}
	if code := commonresponse.CodeFromError(err); code != 401 {
		t.Fatalf("GetMyUser() error code = %d, want 401", code)
	}
}

func TestUpdateMyUserUpdatesEditableFields(t *testing.T) {
	originalNowFunc := nowFunc
	defer func() {
		nowFunc = originalNowFunc
	}()

	fixedTime := time.Date(2026, 4, 20, 12, 30, 0, 0, time.UTC)
	nowFunc = func() time.Time {
		return fixedTime
	}

	loginUser := &model.User{
		Id:          7,
		UserName:    "旧昵称",
		UserAvatar:  "https://example.com/old.png",
		UserProfile: "旧简介",
		UserRole:    "user",
		EditTime:    time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
		CreateTime:  time.Date(2026, 4, 20, 9, 0, 0, 0, time.UTC),
		UpdateTime:  time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
	}
	userModel := &stubUserModel{activeUser: loginUser}
	picturesModel := &stubPicturesModel{
		stats: &model.PictureStats{
			Total:         3,
			ApprovedCount: 2,
			PendingCount:  1,
			RejectedCount: 0,
		},
	}

	logic := NewUpdateMyUserLogic(context.Background(), newUserTestServiceContext("secret", userModel, picturesModel))
	userName := "新昵称"
	userAvatar := "https://example.com/new.png"
	userProfile := "新简介"
	resp, err := logic.UpdateMyUser(&types.UpdateMyUserRequest{
		UserName:    &userName,
		UserAvatar:  &userAvatar,
		UserProfile: &userProfile,
	}, bearerToken(t, "secret", 7))
	if err != nil {
		t.Fatalf("UpdateMyUser() unexpected error = %v", err)
	}
	if userModel.updated == nil {
		t.Fatalf("UpdateMyUser() did not persist user update")
	}
	if userModel.updated.UserName != userName || userModel.updated.UserAvatar != userAvatar || userModel.updated.UserProfile != userProfile {
		t.Fatalf("updated user = %+v", userModel.updated)
	}
	if !userModel.updated.EditTime.Equal(fixedTime) || !userModel.updated.UpdateTime.Equal(fixedTime) {
		t.Fatalf("updated timestamps = edit:%v update:%v, want %v", userModel.updated.EditTime, userModel.updated.UpdateTime, fixedTime)
	}
	if resp == nil || resp.UserName != userName || resp.PictureCount != 3 {
		t.Fatalf("UpdateMyUser() response = %+v", resp)
	}
}

func newUserTestServiceContext(secret string, userModel model.UserModel, picturesModel model.PicturesModel) *svc.ServiceContext {
	cfg := config.Config{}
	cfg.Auth.AccessSecret = secret

	return &svc.ServiceContext{
		Config:        cfg,
		UserModel:     userModel,
		PicturesModel: picturesModel,
	}
}

func bearerToken(t *testing.T, secret string, userID int64) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": userID,
	})
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	return "Bearer " + signed
}

type stubUserModel struct {
	activeUser *model.User
	updated    *model.User
	err        error
}

func (m *stubUserModel) Insert(context.Context, *model.User) (sql.Result, error) {
	panic("unexpected Insert call")
}

func (m *stubUserModel) FindOne(context.Context, int64) (*model.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.activeUser, nil
}

func (m *stubUserModel) FindOneActive(context.Context, int64) (*model.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.activeUser, nil
}

func (m *stubUserModel) FindOneByUserEmail(context.Context, string) (*model.User, error) {
	panic("unexpected FindOneByUserEmail call")
}

func (m *stubUserModel) Update(_ context.Context, data *model.User) error {
	if m.err != nil {
		return m.err
	}
	cloned := *data
	m.updated = &cloned
	return nil
}

func (m *stubUserModel) Delete(context.Context, int64) error {
	panic("unexpected Delete call")
}

func (m *stubUserModel) FindByIDs(context.Context, []int64) ([]*model.User, error) {
	panic("unexpected FindByIDs call")
}

type stubPicturesModel struct {
	stats *model.PictureStats
	err   error
}

func (m *stubPicturesModel) Insert(context.Context, *model.Pictures) (sql.Result, error) {
	panic("unexpected Insert call")
}

func (m *stubPicturesModel) FindOne(context.Context, int64) (*model.Pictures, error) {
	panic("unexpected FindOne call")
}

func (m *stubPicturesModel) Update(context.Context, *model.Pictures) error {
	panic("unexpected Update call")
}

func (m *stubPicturesModel) Delete(context.Context, int64) error {
	panic("unexpected Delete call")
}

func (m *stubPicturesModel) FindOneActive(context.Context, int64) (*model.Pictures, error) {
	panic("unexpected FindOneActive call")
}

func (m *stubPicturesModel) IncrementViewCount(context.Context, int64) error {
	panic("unexpected IncrementViewCount call")
}

func (m *stubPicturesModel) CountByWhere(context.Context, string, ...any) (int64, error) {
	panic("unexpected CountByWhere call")
}

func (m *stubPicturesModel) FindByWhere(context.Context, string, string, int64, int64, ...any) ([]*model.Pictures, error) {
	panic("unexpected FindByWhere call")
}

func (m *stubPicturesModel) CountStatsByUser(context.Context, int64) (*model.PictureStats, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.stats == nil {
		return &model.PictureStats{}, nil
	}
	return m.stats, nil
}
