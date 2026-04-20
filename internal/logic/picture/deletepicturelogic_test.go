package picture

import (
	"context"
	"testing"
	"time"

	commonresponse "photo-album/internal/common/response"
	"photo-album/internal/types"
	"photo-album/model"
)

func TestDeletePictureMarksDeletedForOwner(t *testing.T) {
	t.Parallel()

	const secret = "delete-secret"
	loginUser := &model.User{Id: 7, UserRole: "user"}
	originalTime := time.Now().Add(-time.Hour).UTC()
	picturesModel := &stubPicturesModel{
		picture: &model.Pictures{
			Id:         101,
			Url:        "https://example.com/public/7/demo.jpg",
			Name:       "demo",
			UserId:     loginUser.Id,
			CreateTime: originalTime,
			EditTime:   originalTime,
			UpdateTime: originalTime,
			IsDelete:   0,
		},
	}
	logic := NewDeletePictureLogic(context.Background(), newReviewTestServiceContext(secret, &stubUserModel{user: loginUser}, picturesModel))

	resp, err := logic.DeletePicture(&types.PictureDeleteRequest{Id: 101}, bearerToken(t, secret, loginUser.Id))
	if err != nil {
		t.Fatalf("DeletePicture() unexpected error = %v", err)
	}
	if resp == nil || resp.Id != 101 {
		t.Fatalf("DeletePicture() response = %+v, want id 101", resp)
	}

	if picturesModel.updated == nil {
		t.Fatalf("DeletePicture() did not update picture")
	}
	if picturesModel.updated.IsDelete != 1 {
		t.Fatalf("updated isDelete = %d, want 1", picturesModel.updated.IsDelete)
	}
	if !picturesModel.updated.UpdateTime.After(originalTime) {
		t.Fatalf("updated updateTime = %v, want after %v", picturesModel.updated.UpdateTime, originalTime)
	}
}

func TestDeletePictureAllowsAdmin(t *testing.T) {
	t.Parallel()

	const secret = "delete-secret"
	loginUser := &model.User{Id: 9, UserRole: "admin"}
	picturesModel := &stubPicturesModel{
		picture: &model.Pictures{
			Id:         101,
			Url:        "https://example.com/public/7/demo.jpg",
			Name:       "demo",
			UserId:     7,
			CreateTime: time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
			EditTime:   time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
			UpdateTime: time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
		},
	}
	logic := NewDeletePictureLogic(context.Background(), newReviewTestServiceContext(secret, &stubUserModel{user: loginUser}, picturesModel))

	if _, err := logic.DeletePicture(&types.PictureDeleteRequest{Id: 101}, bearerToken(t, secret, loginUser.Id)); err != nil {
		t.Fatalf("DeletePicture() unexpected error = %v", err)
	}
	if picturesModel.updated == nil || picturesModel.updated.IsDelete != 1 {
		t.Fatalf("DeletePicture() should soft-delete picture, updated = %+v", picturesModel.updated)
	}
}

func TestDeletePictureRejectsNonOwner(t *testing.T) {
	t.Parallel()

	const secret = "delete-secret"
	loginUser := &model.User{Id: 8, UserRole: "user"}
	picturesModel := &stubPicturesModel{
		picture: &model.Pictures{
			Id:         101,
			Url:        "https://example.com/public/7/demo.jpg",
			Name:       "demo",
			UserId:     7,
			CreateTime: time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
			EditTime:   time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
			UpdateTime: time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
		},
	}
	logic := NewDeletePictureLogic(context.Background(), newReviewTestServiceContext(secret, &stubUserModel{user: loginUser}, picturesModel))

	_, err := logic.DeletePicture(&types.PictureDeleteRequest{Id: 101}, bearerToken(t, secret, loginUser.Id))
	if err == nil {
		t.Fatalf("DeletePicture() expected error")
	}
	if code := commonresponse.CodeFromError(err); code != 403 {
		t.Fatalf("DeletePicture() error code = %d, want 403", code)
	}
	if msg := commonresponse.MessageFromError(err); msg != "无权删除该图片" {
		t.Fatalf("DeletePicture() error message = %q, want %q", msg, "无权删除该图片")
	}
	if picturesModel.updated != nil {
		t.Fatalf("DeletePicture() should not update picture for non-owner")
	}
}

func TestDeletePictureReturnsNotFound(t *testing.T) {
	t.Parallel()

	const secret = "delete-secret"
	loginUser := &model.User{Id: 7, UserRole: "user"}
	picturesModel := &stubPicturesModel{err: model.ErrNotFound}
	logic := NewDeletePictureLogic(context.Background(), newReviewTestServiceContext(secret, &stubUserModel{user: loginUser}, picturesModel))

	_, err := logic.DeletePicture(&types.PictureDeleteRequest{Id: 101}, bearerToken(t, secret, loginUser.Id))
	if err == nil {
		t.Fatalf("DeletePicture() expected error")
	}
	if code := commonresponse.CodeFromError(err); code != 404 {
		t.Fatalf("DeletePicture() error code = %d, want 404", code)
	}
}
