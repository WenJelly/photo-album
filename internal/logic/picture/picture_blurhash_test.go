package picture

import (
	"database/sql"
	"testing"

	"photo-album/internal/types"
	"photo-album/model"
)

func TestBuildSolidBlurHashFromHexColor(t *testing.T) {
	hash := buildSolidBlurHash("#A1B2C3")

	if hash == "" {
		t.Fatal("expected blur hash")
	}
	if len(hash) != 6 {
		t.Fatalf("expected minimal blur hash length 6, got %d", len(hash))
	}
	if hash[:2] != "00" {
		t.Fatalf("expected one-component blur hash prefix, got %q", hash[:2])
	}
}

func TestBuildSolidBlurHashRejectsInvalidColors(t *testing.T) {
	if hash := buildSolidBlurHash(""); hash != "" {
		t.Fatalf("expected empty hash for empty color, got %q", hash)
	}
	if hash := buildSolidBlurHash("not-a-color"); hash != "" {
		t.Fatalf("expected empty hash for invalid color, got %q", hash)
	}
}

func TestBuildPictureResponseDerivesBlurHashFromPicColor(t *testing.T) {
	resp, err := buildPictureResponseWithUser(&model.Pictures{
		Id:       1,
		Url:      "https://example.com/photo.jpg",
		Name:     "Photo",
		UserId:   2,
		PicColor: sql.NullString{String: "#A1B2C3", Valid: true},
	}, types.UserDetail{}, types.CompressPictureType{})
	if err != nil {
		t.Fatalf("buildPictureResponseWithUser returned error: %v", err)
	}

	if resp.BlurHash == "" {
		t.Fatal("expected blurHash to be derived from picColor")
	}
}
