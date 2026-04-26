package user

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	commonresponse "photo-album/internal/common/response"
	"photo-album/internal/svc"
	"strings"
	"time"
)

const (
	MaxAvatarMultipartMemory   = 8 << 20
	MaxAvatarUploadSize        = 5 << 20
	AvatarUploadRequestTimeout = 30 * time.Second
)

var (
	loadRequiredLoginUserForAvatar = loadRequiredLoginUser
	saveAvatarMultipartFileToTemp  = saveAvatarMultipartImageToTemp
	buildUserAvatarObjectKeyFunc   = buildUserAvatarObjectKey
	uploadAvatarFileToCOS          = putAvatarFileToCOS
	loadUserPictureStatsForAvatar  = loadUserPictureStats
)

func saveAvatarMultipartImageToTemp(file multipart.File, header *multipart.FileHeader) (string, string, func(), error) {
	if header == nil || strings.TrimSpace(header.Filename) == "" {
		return "", "", nil, commonresponse.BadRequest("上传文件不能为空")
	}

	ext := normalizeAvatarExtension(header.Filename)
	if !isAllowedAvatarExtension(ext) {
		return "", "", nil, commonresponse.BadRequest("仅支持 jpg、jpeg、png、webp 图片")
	}
	if header.Size > MaxAvatarUploadSize {
		return "", "", nil, commonresponse.BadRequest("头像图片大小不能超过 5MB")
	}

	tmpFile, err := os.CreateTemp("", "user-avatar-*"+filepath.Ext(header.Filename))
	if err != nil {
		return "", "", nil, commonresponse.InternalServerError("创建临时文件失败")
	}

	cleanup := func() {
		_ = os.Remove(tmpFile.Name())
	}

	written, copyErr := io.Copy(tmpFile, io.LimitReader(file, MaxAvatarUploadSize+1))
	closeErr := tmpFile.Close()
	if copyErr != nil {
		cleanup()
		return "", "", nil, commonresponse.InternalServerError("写入临时文件失败")
	}
	if closeErr != nil {
		cleanup()
		return "", "", nil, commonresponse.InternalServerError("关闭临时文件失败")
	}
	if written > MaxAvatarUploadSize {
		cleanup()
		return "", "", nil, commonresponse.BadRequest("头像图片大小不能超过 5MB")
	}

	return tmpFile.Name(), header.Filename, cleanup, nil
}

func normalizeAvatarExtension(filename string) string {
	return strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
}

func isAllowedAvatarExtension(ext string) bool {
	switch ext {
	case "jpg", "jpeg", "png", "webp":
		return true
	default:
		return false
	}
}

func buildUserAvatarObjectKey(userID int64, ext string) (string, error) {
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	if ext == "" {
		ext = "jpg"
	}
	if ext == "jpeg" {
		ext = "jpg"
	}

	return fmt.Sprintf("avatar/%d/%s_%s.%s", userID, time.Now().Format("2006-01-02"), hex.EncodeToString(randomBytes), ext), nil
}

func avatarContentType(ext string) string {
	switch ext {
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

func buildAvatarURL(host, objectKey string) string {
	return strings.TrimRight(host, "/") + "/" + escapeAvatarObjectKey(objectKey)
}

func escapeAvatarObjectKey(objectKey string) string {
	parts := strings.Split(strings.TrimLeft(objectKey, "/"), "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}

func putAvatarFileToCOS(ctx context.Context, svcCtx *svc.ServiceContext, localPath, objectKey, contentType string) error {
	if !hasCompleteAvatarCOSConfig(svcCtx) {
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

	targetURL := buildAvatarURL(svcCtx.Config.Cos.Host, objectKey)
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
	req.Header.Set("Authorization", buildAvatarCOSAuthorization(svcCtx.Config.Cos.SecretId, svcCtx.Config.Cos.SecretKey, parsedURL, http.MethodPut))
	req.Host = parsedURL.Host

	resp, err := (&http.Client{Timeout: AvatarUploadRequestTimeout}).Do(req)
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

func hasCompleteAvatarCOSConfig(svcCtx *svc.ServiceContext) bool {
	return strings.TrimSpace(svcCtx.Config.Cos.Host) != "" &&
		strings.TrimSpace(svcCtx.Config.Cos.SecretId) != "" &&
		strings.TrimSpace(svcCtx.Config.Cos.SecretKey) != "" &&
		!strings.Contains(svcCtx.Config.Cos.SecretId, "[REDACTED") &&
		!strings.Contains(svcCtx.Config.Cos.SecretKey, "[REDACTED")
}

func buildAvatarCOSAuthorization(secretID, secretKey string, parsedURL *url.URL, method string) string {
	signTime := fmt.Sprintf("%d;%d", time.Now().Unix()-60, time.Now().Add(10*time.Minute).Unix())
	httpString := fmt.Sprintf("%s\n%s\n\nhost=%s\n", strings.ToLower(method), parsedURL.EscapedPath(), strings.ToLower(parsedURL.Host))
	stringToSign := fmt.Sprintf("sha1\n%s\n%s\n", signTime, sha1HexAvatar(httpString))
	signKey := hmacSha1HexAvatar(secretKey, signTime)
	signature := hmacSha1HexAvatar(signKey, stringToSign)

	return fmt.Sprintf(
		"q-sign-algorithm=sha1&q-ak=%s&q-sign-time=%s&q-key-time=%s&q-header-list=host&q-url-param-list=&q-signature=%s",
		secretID,
		signTime,
		signTime,
		signature,
	)
}

func sha1HexAvatar(value string) string {
	sum := sha1.Sum([]byte(value))
	return hex.EncodeToString(sum[:])
}

func hmacSha1HexAvatar(key, value string) string {
	mac := hmac.New(sha1.New, []byte(key))
	_, _ = mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil))
}
