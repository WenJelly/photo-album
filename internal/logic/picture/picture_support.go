package picture

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	commonresponse "photo-album/internal/common/response"
	"photo-album/internal/svc"
	"photo-album/internal/types"
	"photo-album/model"

	"github.com/golang-jwt/jwt/v4"
)

const (
	// MaxMultipartMemory controls how much multipart form data stays in memory before spilling to disk.
	MaxMultipartMemory = 32 << 20
	// MaxFileUploadSize is the maximum accepted file upload size in bytes.
	MaxFileUploadSize = 30 << 20

	maxURLUploadSize = 10 << 20

	reviewStatusPending = 0
	reviewStatusPass    = 1
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

func isAllowedPictureFilename(filename string) bool {
	switch normalizeExtension(filename) {
	case "jpg", "jpeg", "png", "webp":
		return true
	default:
		return false
	}
}

func normalizeExtension(filename string) string {
	return strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
}

func buildProcessedPictureURLs(host, objectKey string) (string, string) {
	baseURL := buildObjectURL(host, objectKey)
	return baseURL + "?imageMogr2/format/webp", baseURL + "?imageMogr2/thumbnail/128x128%3E/format/webp"
}

func buildObjectURL(host, objectKey string) string {
	return strings.TrimRight(host, "/") + "/" + escapeObjectKey(objectKey)
}

func escapeObjectKey(objectKey string) string {
	parts := strings.Split(strings.TrimLeft(objectKey, "/"), "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}

func extractUserIDFromBearerToken(authorization, secret string) (int64, error) {
	if strings.TrimSpace(secret) == "" {
		return 0, errors.New("jwt secret is empty")
	}

	if authorization == "" {
		return 0, errors.New("missing authorization header")
	}

	parts := strings.SplitN(strings.TrimSpace(authorization), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		return 0, errors.New("invalid authorization header")
	}

	token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || token == nil || !token.Valid {
		return 0, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid token claims")
	}

	return claimToInt64(claims["userId"])
}

func claimToInt64(value any) (int64, error) {
	switch v := value.(type) {
	case float64:
		return int64(v), nil
	case float32:
		return int64(v), nil
	case int64:
		return v, nil
	case int32:
		return int64(v), nil
	case int:
		return int64(v), nil
	case json.Number:
		return v.Int64()
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, errors.New("invalid userId claim")
	}
}

func loadRequiredLoginUser(ctx context.Context, svcCtx *svc.ServiceContext, authorization string) (*model.User, error) {
	userID, err := extractUserIDFromBearerToken(authorization, svcCtx.Config.Auth.AccessSecret)
	if err != nil {
		return nil, commonresponse.Unauthorized("请先登录")
	}

	loginUser, err := svcCtx.UserModel.FindOne(ctx, userID)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, commonresponse.Unauthorized("登录用户不存在")
		}
		return nil, commonresponse.InternalServerError("查询登录用户失败")
	}

	if loginUser.IsDelete != 0 {
		return nil, commonresponse.Unauthorized("登录用户不存在")
	}

	return loginUser, nil
}

func loadOptionalLoginUser(ctx context.Context, svcCtx *svc.ServiceContext, authorization string) (*model.User, error) {
	if strings.TrimSpace(authorization) == "" {
		return nil, nil
	}

	return loadRequiredLoginUser(ctx, svcCtx, authorization)
}

func saveMultipartFileToTemp(file multipart.File, header *multipart.FileHeader) (string, string, func(), error) {
	if header == nil || header.Filename == "" {
		return "", "", nil, commonresponse.BadRequest("上传文件不能为空")
	}
	if !isAllowedPictureFilename(header.Filename) {
		return "", "", nil, commonresponse.BadRequest("仅支持 jpg、jpeg、png、webp 图片")
	}
	if header.Size > MaxFileUploadSize {
		return "", "", nil, commonresponse.BadRequest("图片大小不能超过 30MB")
	}

	tmpFile, err := os.CreateTemp("", "picture-upload-*"+filepath.Ext(header.Filename))
	if err != nil {
		return "", "", nil, commonresponse.InternalServerError("创建临时文件失败")
	}

	cleanup := func() {
		_ = os.Remove(tmpFile.Name())
	}

	written, copyErr := io.Copy(tmpFile, io.LimitReader(file, MaxFileUploadSize+1))
	closeErr := tmpFile.Close()
	if copyErr != nil {
		cleanup()
		return "", "", nil, commonresponse.InternalServerError("写入临时文件失败")
	}
	if closeErr != nil {
		cleanup()
		return "", "", nil, commonresponse.InternalServerError("关闭临时文件失败")
	}
	if written > MaxFileUploadSize {
		cleanup()
		return "", "", nil, commonresponse.BadRequest("图片大小不能超过 30MB")
	}

	return tmpFile.Name(), header.Filename, cleanup, nil
}

func downloadRemoteImageToTemp(ctx context.Context, fileURL string) (string, string, func(), error) {
	parsedURL, err := url.Parse(fileURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return "", "", nil, commonresponse.BadRequest("fileUrl 必须是合法 URL")
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", "", nil, commonresponse.BadRequest("fileUrl 仅支持 http 或 https")
	}

	client := &http.Client{Timeout: 30 * time.Second}

	headReq, _ := http.NewRequestWithContext(ctx, http.MethodHead, fileURL, nil)
	headResp, headErr := client.Do(headReq)
	if headErr == nil && headResp != nil {
		func() {
			defer headResp.Body.Close()
			if headResp.StatusCode == http.StatusOK {
				if contentType := strings.ToLower(headResp.Header.Get("Content-Type")); contentType != "" && !strings.HasPrefix(contentType, "image/") {
					err = commonresponse.BadRequest("远程文件不是图片")
					return
				}
				if length := headResp.ContentLength; length > maxURLUploadSize {
					err = commonresponse.BadRequest("URL 图片大小不能超过 10MB")
					return
				}
			}
		}()
		if err != nil {
			return "", "", nil, err
		}
	}

	getReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, fileURL, nil)
	getResp, err := client.Do(getReq)
	if err != nil {
		return "", "", nil, commonresponse.BadRequest("下载远程图片失败")
	}
	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusOK {
		return "", "", nil, commonresponse.BadRequest("下载远程图片失败")
	}

	if contentType := strings.ToLower(getResp.Header.Get("Content-Type")); contentType != "" && !strings.HasPrefix(contentType, "image/") {
		return "", "", nil, commonresponse.BadRequest("远程文件不是图片")
	}

	originalFilename := deriveRemoteFilename(parsedURL)
	tmpFile, err := os.CreateTemp("", "picture-url-*"+filepath.Ext(originalFilename))
	if err != nil {
		return "", "", nil, commonresponse.InternalServerError("创建临时文件失败")
	}

	cleanup := func() {
		_ = os.Remove(tmpFile.Name())
	}

	written, copyErr := io.Copy(tmpFile, io.LimitReader(getResp.Body, maxURLUploadSize+1))
	closeErr := tmpFile.Close()
	if copyErr != nil {
		cleanup()
		return "", "", nil, commonresponse.InternalServerError("保存远程图片失败")
	}
	if closeErr != nil {
		cleanup()
		return "", "", nil, commonresponse.InternalServerError("关闭临时文件失败")
	}
	if written > maxURLUploadSize {
		cleanup()
		return "", "", nil, commonresponse.BadRequest("URL 图片大小不能超过 10MB")
	}

	return tmpFile.Name(), originalFilename, cleanup, nil
}

func deriveRemoteFilename(parsedURL *url.URL) string {
	name := path.Base(parsedURL.Path)
	name, _ = url.PathUnescape(name)
	if name == "" || name == "." || name == "/" {
		return "remote-image"
	}
	return name
}

func storePicture(ctx context.Context, svcCtx *svc.ServiceContext, tempPath, originalFilename string, req pictureWriteRequest, loginUser *model.User) (*types.PictureResponse, error) {
	existing, err := findPictureForWrite(ctx, svcCtx, req.ID)
	if err != nil {
		return nil, err
	}
	if existing != nil && existing.UserId != loginUser.Id && loginUser.UserRole != "admin" {
		return nil, commonresponse.Forbidden("无权修改该图片")
	}

	metadata, err := extractPictureMetadata(tempPath, originalFilename)
	if err != nil {
		return nil, err
	}

	objectKey, err := buildPictureObjectKey(loginUser.Id, metadata.Format)
	if err != nil {
		return nil, commonresponse.InternalServerError("生成上传路径失败")
	}

	if err := uploadFileToCOS(ctx, svcCtx, tempPath, objectKey, contentTypeForFormat(metadata.Format)); err != nil {
		return nil, err
	}

	now := time.Now()
	objectURL, thumbnailURL := buildProcessedPictureURLs(svcCtx.Config.Cos.Host, objectKey)
	storedPicture := buildPictureModel(existing, req, loginUser, metadata, originalFilename, objectURL, thumbnailURL, now)

	if existing != nil {
		storedPicture.Id = existing.Id
		storedPicture.CreateTime = existing.CreateTime
		storedPicture.UserId = existing.UserId
		storedPicture.ViewCount = existing.ViewCount
		storedPicture.LikeCount = existing.LikeCount

		if err := svcCtx.PicturesModel.Update(ctx, storedPicture); err != nil {
			return nil, commonresponse.InternalServerError("更新图片失败")
		}
		return buildPictureResponseWithUser(storedPicture, buildUserSummary(loginUser)), nil
	}

	result, err := svcCtx.PicturesModel.Insert(ctx, storedPicture)
	if err != nil {
		return nil, commonresponse.InternalServerError("保存图片失败")
	}

	newID, lastIDErr := result.LastInsertId()
	if lastIDErr == nil {
		storedPicture.Id = newID
	}

	return buildPictureResponseWithUser(storedPicture, buildUserSummary(loginUser)), nil
}

func findPictureForWrite(ctx context.Context, svcCtx *svc.ServiceContext, pictureID int64) (*model.Pictures, error) {
	if pictureID <= 0 {
		return nil, nil
	}

	pictureInfo, err := svcCtx.PicturesModel.FindOne(ctx, pictureID)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, commonresponse.NotFound("图片不存在")
		}
		return nil, commonresponse.InternalServerError("查询图片失败")
	}
	if pictureInfo.IsDelete != 0 {
		return nil, commonresponse.NotFound("图片不存在")
	}

	return pictureInfo, nil
}

func buildPictureModel(existing *model.Pictures, req pictureWriteRequest, loginUser *model.User, metadata pictureMetadata, originalFilename, objectURL, thumbnailURL string, now time.Time) *model.Pictures {
	reviewStatus, reviewMessage, reviewerID, reviewTime := reviewStateForUpload(loginUser, now)

	introduction := optionalString(req.Introduction)
	category := optionalString(req.Category)
	tags := tagsJSON(req.Tags)

	if existing != nil {
		if !introduction.Valid {
			introduction = existing.Introduction
		}
		if !category.Valid {
			category = existing.Category
		}
		if !tags.Valid {
			tags = existing.Tags
		}
	}

	return &model.Pictures{
		Url:           objectURL,
		Name:          resolvePictureName(req.PicName, originalFilename),
		Introduction:  introduction,
		Category:      category,
		Tags:          tags,
		PicSize:       optionalInt64(metadata.Size),
		PicWidth:      optionalInt64(metadata.Width),
		PicHeight:     optionalInt64(metadata.Height),
		PicScale:      optionalFloat64(metadata.Scale),
		PicFormat:     optionalString(metadata.Format),
		UserId:        loginUser.Id,
		CreateTime:    now,
		EditTime:      now,
		UpdateTime:    now,
		ReviewStatus:  reviewStatus,
		ReviewMessage: reviewMessage,
		ReviewerId:    reviewerID,
		ReviewTime:    reviewTime,
		ThumbnailUrl:  optionalString(thumbnailURL),
		PicColor:      optionalString(metadata.DominantColor),
		ViewCount:     0,
		LikeCount:     0,
		IsDelete:      0,
	}
}

func reviewStateForUpload(loginUser *model.User, now time.Time) (int64, sql.NullString, sql.NullInt64, sql.NullTime) {
	if loginUser.UserRole == "admin" {
		return reviewStatusPass, optionalString("管理员上传自动通过"), optionalInt64(loginUser.Id), optionalTime(now)
	}

	return reviewStatusPending, optionalString("待审核"), sql.NullInt64{}, sql.NullTime{}
}

func resolvePictureName(picName, originalFilename string) string {
	if trimmed := strings.TrimSpace(picName); trimmed != "" {
		return trimmed
	}

	baseName := filepath.Base(originalFilename)
	ext := filepath.Ext(baseName)
	baseName = strings.TrimSpace(strings.TrimSuffix(baseName, ext))
	if baseName == "" || baseName == "." {
		return "未命名图片"
	}
	return baseName
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

func buildPictureObjectKey(userID int64, format string) (string, error) {
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	ext := format
	if ext == "jpeg" {
		ext = "jpg"
	}
	if ext == "" {
		ext = "jpg"
	}

	return fmt.Sprintf("public/%d/%s_%s.%s", userID, time.Now().Format("2006-01-02"), hex.EncodeToString(randomBytes), ext), nil
}

func contentTypeForFormat(format string) string {
	switch format {
	case "jpg", "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}

func uploadFileToCOS(ctx context.Context, svcCtx *svc.ServiceContext, localPath, objectKey, contentType string) error {
	if strings.TrimSpace(svcCtx.Config.Cos.Host) == "" ||
		strings.TrimSpace(svcCtx.Config.Cos.SecretId) == "" ||
		strings.TrimSpace(svcCtx.Config.Cos.SecretKey) == "" ||
		strings.Contains(svcCtx.Config.Cos.SecretId, "[REDACTED") ||
		strings.Contains(svcCtx.Config.Cos.SecretKey, "[REDACTED") {
		return commonresponse.InternalServerError("COS 配置不完整，请先配置本地密钥")
	}

	file, err := os.Open(localPath)
	if err != nil {
		return commonresponse.InternalServerError("读取上传文件失败")
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return commonresponse.InternalServerError("读取文件信息失败")
	}

	targetURL := buildObjectURL(svcCtx.Config.Cos.Host, objectKey)
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return commonresponse.InternalServerError("生成 COS 地址失败")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, targetURL, file)
	if err != nil {
		return commonresponse.InternalServerError("创建 COS 请求失败")
	}
	req.ContentLength = info.Size()
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", buildCOSAuthorization(svcCtx.Config.Cos.SecretId, svcCtx.Config.Cos.SecretKey, parsedURL))
	req.Host = parsedURL.Host

	resp, err := (&http.Client{Timeout: 60 * time.Second}).Do(req)
	if err != nil {
		return commonresponse.InternalServerError("上传 COS 失败")
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return commonresponse.InternalServerError(fmt.Sprintf("COS 上传失败: %s", strings.TrimSpace(string(body))))
	}

	return nil
}

func buildCOSAuthorization(secretID, secretKey string, parsedURL *url.URL) string {
	signTime := fmt.Sprintf("%d;%d", time.Now().Unix()-60, time.Now().Add(10*time.Minute).Unix())
	httpString := fmt.Sprintf("%s\n%s\n\nhost=%s\n", strings.ToLower(http.MethodPut), parsedURL.EscapedPath(), strings.ToLower(parsedURL.Host))
	stringToSign := fmt.Sprintf("sha1\n%s\n%s\n", signTime, sha1Hex(httpString))
	signKey := hmacSha1Hex(secretKey, signTime)
	signature := hmacSha1Hex(signKey, stringToSign)

	return fmt.Sprintf(
		"q-sign-algorithm=sha1&q-ak=%s&q-sign-time=%s&q-key-time=%s&q-header-list=host&q-url-param-list=&q-signature=%s",
		secretID,
		signTime,
		signTime,
		signature,
	)
}

func sha1Hex(value string) string {
	sum := sha1.Sum([]byte(value))
	return hex.EncodeToString(sum[:])
}

func hmacSha1Hex(key, value string) string {
	mac := hmac.New(sha1.New, []byte(key))
	_, _ = mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil))
}

func extractPictureMetadata(filePath, originalFilename string) (pictureMetadata, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return pictureMetadata{}, commonresponse.InternalServerError("读取图片失败")
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return pictureMetadata{}, commonresponse.InternalServerError("读取图片信息失败")
	}

	header := make([]byte, 64)
	n, _ := file.Read(header)
	header = header[:n]

	format := detectPictureFormat(header, originalFilename)
	if format == "" {
		return pictureMetadata{}, commonresponse.BadRequest("仅支持 jpg、jpeg、png、webp 图片")
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return pictureMetadata{}, commonresponse.InternalServerError("读取图片失败")
	}

	metadata := pictureMetadata{
		Size:   info.Size(),
		Format: format,
	}

	switch format {
	case "jpg", "jpeg", "png":
		cfg, _, err := image.DecodeConfig(file)
		if err != nil {
			return pictureMetadata{}, commonresponse.BadRequest("无法解析图片尺寸")
		}
		metadata.Width = int64(cfg.Width)
		metadata.Height = int64(cfg.Height)
		if metadata.Height > 0 {
			metadata.Scale = float64(metadata.Width) / float64(metadata.Height)
		}

		if colorValue, err := extractDominantColor(filePath); err == nil {
			metadata.DominantColor = colorValue
		}
	case "webp":
		width, height, err := extractWebPDimensions(header)
		if err != nil {
			return pictureMetadata{}, commonresponse.BadRequest("无法解析 webp 图片尺寸")
		}
		metadata.Width = width
		metadata.Height = height
		if metadata.Height > 0 {
			metadata.Scale = float64(metadata.Width) / float64(metadata.Height)
		}
	}

	return metadata, nil
}

func detectPictureFormat(header []byte, originalFilename string) string {
	switch {
	case len(header) >= 3 && bytes.Equal(header[:3], []byte{0xFF, 0xD8, 0xFF}):
		return "jpg"
	case len(header) >= 8 && bytes.Equal(header[:8], []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}):
		return "png"
	case len(header) >= 12 && bytes.Equal(header[:4], []byte("RIFF")) && bytes.Equal(header[8:12], []byte("WEBP")):
		return "webp"
	default:
		ext := normalizeExtension(originalFilename)
		if ext == "jpeg" {
			return "jpeg"
		}
		if ext == "jpg" || ext == "png" || ext == "webp" {
			return ext
		}
		return ""
	}
}

func extractDominantColor(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return "", err
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width == 0 || height == 0 {
		return "", errors.New("invalid image bounds")
	}

	stepX := maxInt(width/32, 1)
	stepY := maxInt(height/32, 1)

	var totalR, totalG, totalB, count uint64
	for y := bounds.Min.Y; y < bounds.Max.Y; y += stepY {
		for x := bounds.Min.X; x < bounds.Max.X; x += stepX {
			r, g, b, _ := img.At(x, y).RGBA()
			totalR += uint64(r >> 8)
			totalG += uint64(g >> 8)
			totalB += uint64(b >> 8)
			count++
		}
	}

	if count == 0 {
		return "", errors.New("empty color sample")
	}

	return fmt.Sprintf("#%02X%02X%02X", totalR/count, totalG/count, totalB/count), nil
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}
