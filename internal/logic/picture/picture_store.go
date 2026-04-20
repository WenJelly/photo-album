package picture

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"time"

	commonresponse "photo-album/internal/common/response"
	"photo-album/internal/svc"
	"photo-album/internal/types"
	"photo-album/model"
)

func storePicture(ctx context.Context, svcCtx *svc.ServiceContext, tempPath, originalFilename string, req pictureWriteRequest, loginUser *model.User) (*types.PictureResponse, error) {
	existing, err := findPictureForWrite(ctx, svcCtx, req.ID)
	if err != nil {
		return nil, err
	}
	if existing != nil && !canManagePicture(existing.UserId, loginUser) {
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
	objectURL, thumbnailURL := buildStoredPictureURLs(svcCtx.Config.Cos.Host, objectKey, metadata.Size)
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
