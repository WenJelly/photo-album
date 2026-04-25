package picture

import "testing"

func TestEncodeDecodePictureCursor(t *testing.T) {
	token, err := encodePictureCursor(12345)
	if err != nil {
		t.Fatalf("encodePictureCursor returned error: %v", err)
	}

	if token == "" {
		t.Fatal("expected non-empty cursor token")
	}

	id, err := decodePictureCursor(token)
	if err != nil {
		t.Fatalf("decodePictureCursor returned error: %v", err)
	}

	if id != 12345 {
		t.Fatalf("unexpected cursor id: %d", id)
	}
}

func TestDecodePictureCursorRejectsInvalidTokens(t *testing.T) {
	if _, err := decodePictureCursor("not-a-cursor"); err == nil {
		t.Fatal("expected invalid cursor to return an error")
	}
}

func TestNormalizePictureCursorPage(t *testing.T) {
	pageSize, err := normalizePictureCursorPage(0)
	if err != nil {
		t.Fatalf("normalizePictureCursorPage returned error: %v", err)
	}
	if pageSize != defaultPictureCursorPageSize {
		t.Fatalf("unexpected default page size: %d", pageSize)
	}

	pageSize, err = normalizePictureCursorPage(maxPictureCursorPageSize)
	if err != nil {
		t.Fatalf("normalizePictureCursorPage returned error for max size: %v", err)
	}
	if pageSize != maxPictureCursorPageSize {
		t.Fatalf("unexpected max page size: %d", pageSize)
	}

	if _, err := normalizePictureCursorPage(maxPictureCursorPageSize + 1); err == nil {
		t.Fatal("expected oversized cursor page to fail")
	}
}
