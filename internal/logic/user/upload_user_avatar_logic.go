package user

import (
	"context"
	"mime/multipart"

	commonresponse "photo-album/internal/common/response"
	"photo-album/internal/svc"
	"photo-album/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UploadUserAvatarLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUploadUserAvatarLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadUserAvatarLogic {
	return &UploadUserAvatarLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadUserAvatarLogic) UploadUserAvatar(file multipart.File, header *multipart.FileHeader, authorization string) (*types.DetailUserResponse, error) {
	loginUser, err := loadRequiredLoginUserForAvatar(l.ctx, l.svcCtx, authorization)
	if err != nil {
		return nil, err
	}

	tempPath, originalFilename, cleanup, err := saveAvatarMultipartFileToTemp(file, header)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	ext := normalizeAvatarExtension(originalFilename)
	objectKey, err := buildUserAvatarObjectKeyFunc(loginUser.Id, ext)
	if err != nil {
		return nil, commonresponse.InternalServerError("生成头像上传路径失败")
	}

	if err := uploadAvatarFileToCOS(l.ctx, l.svcCtx, tempPath, objectKey, avatarContentType(ext)); err != nil {
		return nil, err
	}

	updatedUser := cloneUser(loginUser)
	now := nowFunc()
	updatedUser.UserAvatar = buildAvatarURL(l.svcCtx.Config.Cos.Host, objectKey)
	updatedUser.EditTime = now
	updatedUser.UpdateTime = now

	if err := l.svcCtx.UserModel.Update(l.ctx, updatedUser); err != nil {
		return nil, commonresponse.InternalServerError("更新用户头像失败")
	}

	pictureStats, err := loadUserPictureStatsForAvatar(l.ctx, l.svcCtx, updatedUser.Id)
	if err != nil {
		return nil, err
	}

	return buildDetailUserResponse(updatedUser, pictureStats), nil
}
