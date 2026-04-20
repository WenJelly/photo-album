package picture

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

func TestValidateReviewDecision(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		status      int64
		message     string
		wantErrCode int
		wantErrMsg  string
	}{
		{
			name:   "pass allows empty message",
			status: reviewStatusPass,
		},
		{
			name:        "reject requires message",
			status:      reviewStatusReject,
			message:     "   ",
			wantErrCode: 400,
			wantErrMsg:  "reviewMessage 不能为空",
		},
		{
			name:    "reject allows non-empty message",
			status:  reviewStatusReject,
			message: "图片内容不合规",
		},
		{
			name:        "invalid status rejected",
			status:      reviewStatusPending,
			message:     "ignored",
			wantErrCode: 400,
			wantErrMsg:  "reviewStatus 只能是 1 或 2",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateReviewDecision(tt.status, tt.message)
			if tt.wantErrCode == 0 {
				if err != nil {
					t.Fatalf("validateReviewDecision() unexpected error = %v", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("validateReviewDecision() expected error")
			}
			if code := commonresponse.CodeFromError(err); code != tt.wantErrCode {
				t.Fatalf("validateReviewDecision() code = %d, want %d", code, tt.wantErrCode)
			}
			if msg := commonresponse.MessageFromError(err); msg != tt.wantErrMsg {
				t.Fatalf("validateReviewDecision() message = %q, want %q", msg, tt.wantErrMsg)
			}
		})
	}
}

func TestReviewPictureUpdatesAuditFields(t *testing.T) {
	t.Parallel()

	const secret = "review-secret"
	admin := &model.User{Id: 9, UserRole: "admin"}
	originalTime := time.Now().Add(-time.Hour).UTC()
	pictureInfo := &model.Pictures{
		Id:           101,
		Url:          "https://example.com/demo.webp",
		Name:         "demo",
		UserId:       7,
		CreateTime:   originalTime,
		EditTime:     originalTime,
		UpdateTime:   originalTime,
		ReviewStatus: reviewStatusPending,
	}

	userModel := &stubUserModel{user: admin}
	picturesModel := &stubPicturesModel{picture: pictureInfo}
	logic := NewReviewPictureLogic(context.Background(), newReviewTestServiceContext(secret, userModel, picturesModel))

	resp, err := logic.ReviewPicture(&types.PictureReviewRequest{
		Id:            types.NewSnowflakeID(pictureInfo.Id),
		ReviewStatus:  reviewStatusPass,
		ReviewMessage: "审核通过",
	}, bearerToken(t, secret, admin.Id))
	if err != nil {
		t.Fatalf("ReviewPicture() unexpected error = %v", err)
	}
	if resp == nil {
		t.Fatalf("ReviewPicture() returned nil response")
	}
	if resp.ReviewStatus != reviewStatusPass {
		t.Fatalf("ReviewPicture() reviewStatus = %d, want %d", resp.ReviewStatus, reviewStatusPass)
	}
	if resp.ReviewerId.Int64() != admin.Id {
		t.Fatalf("ReviewPicture() reviewerId = %d, want %d", resp.ReviewerId, admin.Id)
	}
	if resp.ReviewMessage != "审核通过" {
		t.Fatalf("ReviewPicture() reviewMessage = %q, want %q", resp.ReviewMessage, "审核通过")
	}

	if picturesModel.updated == nil {
		t.Fatalf("ReviewPicture() did not update picture")
	}
	if picturesModel.updated.ReviewStatus != reviewStatusPass {
		t.Fatalf("updated reviewStatus = %d, want %d", picturesModel.updated.ReviewStatus, reviewStatusPass)
	}
	if !picturesModel.updated.ReviewerId.Valid || picturesModel.updated.ReviewerId.Int64 != admin.Id {
		t.Fatalf("updated reviewerId = %+v, want %d", picturesModel.updated.ReviewerId, admin.Id)
	}
	if !picturesModel.updated.ReviewTime.Valid {
		t.Fatalf("updated reviewTime should be set")
	}
	if !picturesModel.updated.UpdateTime.After(originalTime) {
		t.Fatalf("updated updateTime = %v, want after %v", picturesModel.updated.UpdateTime, originalTime)
	}
}

func TestReviewPictureRejectRequiresMessage(t *testing.T) {
	t.Parallel()

	const secret = "review-secret"
	admin := &model.User{Id: 9, UserRole: "admin"}
	userModel := &stubUserModel{user: admin}
	picturesModel := &stubPicturesModel{
		picture: &model.Pictures{
			Id:           101,
			Url:          "https://example.com/demo.webp",
			Name:         "demo",
			UserId:       7,
			CreateTime:   time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
			EditTime:     time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
			UpdateTime:   time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
			ReviewStatus: reviewStatusPending,
		},
	}
	logic := NewReviewPictureLogic(context.Background(), newReviewTestServiceContext(secret, userModel, picturesModel))

	_, err := logic.ReviewPicture(&types.PictureReviewRequest{
		Id:            101,
		ReviewStatus:  reviewStatusReject,
		ReviewMessage: "   ",
	}, bearerToken(t, secret, admin.Id))
	if err == nil {
		t.Fatalf("ReviewPicture() expected error")
	}
	if code := commonresponse.CodeFromError(err); code != 400 {
		t.Fatalf("ReviewPicture() error code = %d, want 400", code)
	}
	if msg := commonresponse.MessageFromError(err); msg != "reviewMessage 不能为空" {
		t.Fatalf("ReviewPicture() error message = %q, want %q", msg, "reviewMessage 不能为空")
	}
	if picturesModel.updated != nil {
		t.Fatalf("ReviewPicture() should not update picture on validation error")
	}
}

func TestReviewPictureRejectsNonAdmin(t *testing.T) {
	t.Parallel()

	const secret = "review-secret"
	loginUser := &model.User{Id: 8, UserRole: "user"}
	userModel := &stubUserModel{user: loginUser}
	picturesModel := &stubPicturesModel{
		picture: &model.Pictures{
			Id:           101,
			Url:          "https://example.com/demo.webp",
			Name:         "demo",
			UserId:       7,
			CreateTime:   time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
			EditTime:     time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
			UpdateTime:   time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
			ReviewStatus: reviewStatusPending,
		},
	}
	logic := NewReviewPictureLogic(context.Background(), newReviewTestServiceContext(secret, userModel, picturesModel))

	_, err := logic.ReviewPicture(&types.PictureReviewRequest{
		Id:           101,
		ReviewStatus: reviewStatusPass,
	}, bearerToken(t, secret, loginUser.Id))
	if err == nil {
		t.Fatalf("ReviewPicture() expected error")
	}
	if code := commonresponse.CodeFromError(err); code != 403 {
		t.Fatalf("ReviewPicture() error code = %d, want 403", code)
	}
	if picturesModel.updated != nil {
		t.Fatalf("ReviewPicture() should not update picture for non-admin user")
	}
}

func newReviewTestServiceContext(secret string, userModel model.UserModel, picturesModel model.PicturesModel) *svc.ServiceContext {
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
	user *model.User
	err  error
}

func (m *stubUserModel) Insert(context.Context, *model.User) (sql.Result, error) {
	panic("unexpected Insert call")
}

func (m *stubUserModel) FindOne(context.Context, int64) (*model.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.user, nil
}

func (m *stubUserModel) FindOneActive(context.Context, int64) (*model.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.user, nil
}

func (m *stubUserModel) FindOneByUserEmail(context.Context, string) (*model.User, error) {
	panic("unexpected FindOneByUserEmail call")
}

func (m *stubUserModel) Update(context.Context, *model.User) error {
	panic("unexpected Update call")
}

func (m *stubUserModel) Delete(context.Context, int64) error {
	panic("unexpected Delete call")
}

func (m *stubUserModel) FindByIDs(context.Context, []int64) ([]*model.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.user == nil {
		return nil, nil
	}
	return []*model.User{m.user}, nil
}

type stubPicturesModel struct {
	picture *model.Pictures
	err     error
	updated *model.Pictures
}

func (m *stubPicturesModel) Insert(context.Context, *model.Pictures) (sql.Result, error) {
	panic("unexpected Insert call")
}

func (m *stubPicturesModel) FindOne(context.Context, int64) (*model.Pictures, error) {
	panic("unexpected FindOne call")
}

func (m *stubPicturesModel) Update(_ context.Context, picture *model.Pictures) error {
	if m.err != nil {
		return m.err
	}
	cloned := *picture
	m.updated = &cloned
	return nil
}

func (m *stubPicturesModel) Delete(context.Context, int64) error {
	panic("unexpected Delete call")
}

func (m *stubPicturesModel) FindOneActive(context.Context, int64) (*model.Pictures, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.picture, nil
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
	panic("unexpected CountStatsByUser call")
}
