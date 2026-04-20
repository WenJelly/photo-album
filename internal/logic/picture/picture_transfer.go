package picture

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	commonresponse "photo-album/internal/common/response"
)

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
