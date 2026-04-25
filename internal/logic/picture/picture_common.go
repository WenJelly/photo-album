package picture

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"
)

const (
	// MaxMultipartMemory controls how much multipart form data stays in memory before spilling to disk.
	MaxMultipartMemory = 32 << 20
	// MaxFileUploadSize is the maximum accepted file upload size in bytes.
	MaxFileUploadSize = 30 << 20

	maxURLUploadSize          = 10 << 20
	compressedImageThreshold  = 2 << 20
	mediumCompressedThreshold = 5 << 20
	largeCompressedThreshold  = 10 << 20

	reviewStatusPending = 0
	reviewStatusPass    = 1
	reviewStatusReject  = 2
)

type pictureWriteRequest struct {
	ID           int64
	PicName      string
	Introduction string
	Category     string
	Tags         []string
}

type pictureMetadata struct {
	Size          int64
	Width         int64
	Height        int64
	Scale         float64
	Format        string
	DominantColor string
	BlurHash      string
}

func ParseTagsInput(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	if strings.HasPrefix(raw, "[") {
		var tags []string
		if err := json.Unmarshal([]byte(raw), &tags); err != nil {
			return nil, err
		}
		return normalizeTags(tags), nil
	}

	return normalizeTags(strings.Split(raw, ",")), nil
}

func normalizeTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(tags))
	normalized := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		normalized = append(normalized, tag)
	}

	if len(normalized) == 0 {
		return nil
	}

	return normalized
}

func parseStoredTags(raw sql.NullString) []string {
	if !raw.Valid || strings.TrimSpace(raw.String) == "" {
		return nil
	}

	var tags []string
	if err := json.Unmarshal([]byte(raw.String), &tags); err != nil {
		return normalizeTags(strings.Split(raw.String, ","))
	}
	return normalizeTags(tags)
}

func tagsJSON(tags []string) sql.NullString {
	tags = normalizeTags(tags)
	if len(tags) == 0 {
		return sql.NullString{}
	}

	data, err := json.Marshal(tags)
	if err != nil {
		return sql.NullString{}
	}

	return sql.NullString{String: string(data), Valid: true}
}

func optionalString(value string) sql.NullString {
	value = strings.TrimSpace(value)
	if value == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: value, Valid: true}
}

func optionalInt64(value int64) sql.NullInt64 {
	if value <= 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: value, Valid: true}
}

func optionalFloat64(value float64) sql.NullFloat64 {
	if value <= 0 {
		return sql.NullFloat64{}
	}
	return sql.NullFloat64{Float64: value, Valid: true}
}

func optionalTime(value time.Time) sql.NullTime {
	if value.IsZero() {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: value, Valid: true}
}

func nullStringValue(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func nullInt64Value(value sql.NullInt64) int64 {
	if !value.Valid {
		return 0
	}
	return value.Int64
}

func nullFloat64Value(value sql.NullFloat64) float64 {
	if !value.Valid {
		return 0
	}
	return value.Float64
}

func nullTimeValue(value sql.NullTime) string {
	if !value.Valid {
		return ""
	}
	return value.Time.Format("2006-01-02 15:04:05")
}
