package picture

import "testing"

func TestBuildStoredPictureURLs(t *testing.T) {
	t.Parallel()

	const (
		host      = "https://example.com"
		objectKey = "public/1/demo.jpg"
	)

	tests := []struct {
		name             string
		size             int64
		wantURL          string
		wantThumbnailURL string
	}{
		{
			name:             "smaller than threshold keeps original urls",
			size:             2<<20 - 1,
			wantURL:          "https://example.com/public/1/demo.jpg",
			wantThumbnailURL: "https://example.com/public/1/demo.jpg",
		},
		{
			name:             "exactly 2mb keeps original urls",
			size:             2 << 20,
			wantURL:          "https://example.com/public/1/demo.jpg",
			wantThumbnailURL: "https://example.com/public/1/demo.jpg",
		},
		{
			name:             "greater than threshold uses compressed image url",
			size:             2<<20 + 1,
			wantURL:          "https://example.com/public/1/demo.jpg",
			wantThumbnailURL: "https://example.com/public/1/demo.jpg?imageMogr2/thumbnail/2560x2560>/format/webp/quality/85!/minsize/1/ignore-error/1",
		},
		{
			name:             "larger files use a tighter quality tier",
			size:             5<<20 + 1,
			wantURL:          "https://example.com/public/1/demo.jpg",
			wantThumbnailURL: "https://example.com/public/1/demo.jpg?imageMogr2/thumbnail/1920x1920>/format/webp/quality/80!/minsize/1/ignore-error/1",
		},
		{
			name:             "very large files use the strongest thumbnail tier",
			size:             10 << 20,
			wantURL:          "https://example.com/public/1/demo.jpg",
			wantThumbnailURL: "https://example.com/public/1/demo.jpg?imageMogr2/thumbnail/1600x1600>/format/webp/quality/75!/minsize/1/ignore-error/1",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotURL, gotThumbnailURL := buildStoredPictureURLs(host, objectKey, tt.size)
			if gotURL != tt.wantURL {
				t.Fatalf("buildStoredPictureURLs() url = %q, want %q", gotURL, tt.wantURL)
			}
			if gotThumbnailURL != tt.wantThumbnailURL {
				t.Fatalf("buildStoredPictureURLs() thumbnailUrl = %q, want %q", gotThumbnailURL, tt.wantThumbnailURL)
			}
		})
	}
}

func TestExtractObjectKeyFromURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		host    string
		fileURL string
		wantKey string
		wantOK  bool
	}{
		{
			name:    "matches direct host path",
			host:    "https://example.com",
			fileURL: "https://example.com/public/1/demo.jpg",
			wantKey: "public/1/demo.jpg",
			wantOK:  true,
		},
		{
			name:    "supports host base path and escaped filename",
			host:    "https://example.com/assets",
			fileURL: "https://example.com/assets/public/1/demo%20image.jpg",
			wantKey: "public/1/demo image.jpg",
			wantOK:  true,
		},
		{
			name:    "rejects different host",
			host:    "https://example.com",
			fileURL: "https://cdn.example.com/public/1/demo.jpg",
			wantKey: "",
			wantOK:  false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotKey, gotOK := extractObjectKeyFromURL(tt.host, tt.fileURL)
			if gotOK != tt.wantOK {
				t.Fatalf("extractObjectKeyFromURL() ok = %v, want %v", gotOK, tt.wantOK)
			}
			if gotKey != tt.wantKey {
				t.Fatalf("extractObjectKeyFromURL() key = %q, want %q", gotKey, tt.wantKey)
			}
		})
	}
}
