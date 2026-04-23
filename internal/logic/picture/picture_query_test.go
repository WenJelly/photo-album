package picture

import (
	"testing"

	"photo-album/internal/types"
)

func TestBuildPictureListWhere_UsesPublicReviewFilter(t *testing.T) {
	whereSQL, args, err := buildPictureListWhere(&types.QueryPictureRequest{}, true)
	if err != nil {
		t.Fatalf("buildPictureListWhere returned error: %v", err)
	}

	if whereSQL != "where `isDelete` = 0 and `reviewStatus` = ?" {
		t.Fatalf("unexpected where sql: %s", whereSQL)
	}

	if len(args) != 1 || args[0] != int64(reviewStatusPass) {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestBuildAdminPictureListWhere_OmitsReviewFilterWhenUnset(t *testing.T) {
	whereSQL, args, err := buildAdminPictureListWhere(&types.AdminQueryPictureRequest{ReviewStatus: -1})
	if err != nil {
		t.Fatalf("buildAdminPictureListWhere returned error: %v", err)
	}

	if whereSQL != "where `isDelete` = 0" {
		t.Fatalf("unexpected where sql: %s", whereSQL)
	}

	if len(args) != 0 {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestBuildAdminPictureListWhere_FiltersPendingStatus(t *testing.T) {
	whereSQL, args, err := buildAdminPictureListWhere(&types.AdminQueryPictureRequest{ReviewStatus: reviewStatusPending})
	if err != nil {
		t.Fatalf("buildAdminPictureListWhere returned error: %v", err)
	}

	expectedSQL := "where `isDelete` = 0 and `reviewStatus` = ?"
	if whereSQL != expectedSQL {
		t.Fatalf("unexpected where sql: %s", whereSQL)
	}

	if len(args) != 1 || args[0] != int64(reviewStatusPending) {
		t.Fatalf("unexpected args: %#v", args)
	}
}
