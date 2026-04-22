package picture

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"photo-album/internal/types"
	"photo-album/model"
)

func TestListMyPicturesForcesCurrentUserID(t *testing.T) {
	const secret = "list-secret"

	loginUser := &model.User{
		Id:         7,
		UserName:   "当前用户",
		UserRole:   "user",
		CreateTime: time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
		UpdateTime: time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
	}
	picturesModel := &listStubPicturesModel{
		total: 1,
		pictures: []*model.Pictures{
			{
				Id:           101,
				Url:          "https://example.com/demo.webp",
				Name:         "demo",
				UserId:       loginUser.Id,
				CreateTime:   time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
				EditTime:     time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
				UpdateTime:   time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
				ReviewStatus: reviewStatusPending,
			},
		},
	}

	logic := NewListPictureLogic(context.Background(), newReviewTestServiceContext(secret, &stubUserModel{user: loginUser}, picturesModel))
	req := &types.PictureListRequest{
		PageNum:  1,
		PageSize: 10,
		UserId:   types.NewSnowflakeID(999),
	}
	resp, err := logic.ListMyPictures(req, bearerToken(t, secret, loginUser.Id))
	if err != nil {
		t.Fatalf("ListMyPictures() unexpected error = %v", err)
	}
	if resp == nil || resp.Total != 1 || len(resp.List) != 1 {
		t.Fatalf("ListMyPictures() response = %+v", resp)
	}
	if !strings.Contains(picturesModel.lastWhereSQL, "`userId` = ?") {
		t.Fatalf("ListMyPictures() whereSQL = %q, want userId filter", picturesModel.lastWhereSQL)
	}
	if len(picturesModel.lastArgs) == 0 {
		t.Fatalf("ListMyPictures() args should contain current user id")
	}
	found := false
	for _, arg := range picturesModel.lastArgs {
		if value, ok := arg.(int64); ok && value == loginUser.Id {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("ListMyPictures() args = %#v, want current user id %d", picturesModel.lastArgs, loginUser.Id)
	}
	if resp.List[0].UserId.Int64() != loginUser.Id {
		t.Fatalf("ListMyPictures() userId = %d, want %d", resp.List[0].UserId, loginUser.Id)
	}
}

func TestListPictureRawIncludesUserSummary(t *testing.T) {
	const secret = "list-secret"

	adminUser := &model.User{
		Id:         9,
		UserName:   "管理员",
		UserAvatar: "https://example.com/admin.png",
		UserRole:   "admin",
		CreateTime: time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
		UpdateTime: time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
	}
	picturesModel := &listStubPicturesModel{
		total: 1,
		pictures: []*model.Pictures{
			{
				Id:           101,
				Url:          "https://example.com/demo.webp",
				Name:         "demo",
				UserId:       adminUser.Id,
				CreateTime:   time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
				EditTime:     time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
				UpdateTime:   time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
				ReviewStatus: reviewStatusPending,
			},
		},
	}

	logic := NewListPictureLogic(context.Background(), newReviewTestServiceContext(secret, &stubUserModel{user: adminUser}, picturesModel))
	resp, err := logic.ListPictureRaw(&types.PictureListRequest{PageNum: 1, PageSize: 10}, bearerToken(t, secret, adminUser.Id))
	if err != nil {
		t.Fatalf("ListPictureRaw() unexpected error = %v", err)
	}
	if resp == nil || len(resp.List) != 1 {
		t.Fatalf("ListPictureRaw() response = %+v", resp)
	}
	if resp.List[0].User == nil {
		t.Fatalf("ListPictureRaw() user summary is nil")
	}
	if resp.List[0].User.Id.Int64() != adminUser.Id {
		t.Fatalf("ListPictureRaw() user id = %d, want %d", resp.List[0].User.Id, adminUser.Id)
	}
	if resp.List[0].User.UserName != adminUser.UserName {
		t.Fatalf("ListPictureRaw() userName = %q, want %q", resp.List[0].User.UserName, adminUser.UserName)
	}
}

type listStubPicturesModel struct {
	total        int64
	pictures     []*model.Pictures
	lastWhereSQL string
	lastArgs     []any
}

func (m *listStubPicturesModel) Insert(context.Context, *model.Pictures) (sql.Result, error) {
	panic("unexpected Insert call")
}

func (m *listStubPicturesModel) FindOne(context.Context, int64) (*model.Pictures, error) {
	panic("unexpected FindOne call")
}

func (m *listStubPicturesModel) Update(context.Context, *model.Pictures) error {
	panic("unexpected Update call")
}

func (m *listStubPicturesModel) Delete(context.Context, int64) error {
	panic("unexpected Delete call")
}

func (m *listStubPicturesModel) FindOneActive(context.Context, int64) (*model.Pictures, error) {
	panic("unexpected FindOneActive call")
}

func (m *listStubPicturesModel) IncrementViewCount(context.Context, int64) error {
	panic("unexpected IncrementViewCount call")
}

func (m *listStubPicturesModel) CountByWhere(_ context.Context, whereSQL string, args ...any) (int64, error) {
	m.lastWhereSQL = whereSQL
	m.lastArgs = append([]any{}, args...)
	return m.total, nil
}

func (m *listStubPicturesModel) FindByWhere(_ context.Context, whereSQL, _ string, _, _ int64, args ...any) ([]*model.Pictures, error) {
	m.lastWhereSQL = whereSQL
	m.lastArgs = append([]any{}, args...)
	return m.pictures, nil
}

func (m *listStubPicturesModel) CountStatsByUser(context.Context, int64) (*model.PictureStats, error) {
	panic("unexpected CountStatsByUser call")
}
