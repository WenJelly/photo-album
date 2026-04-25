package picture

import (
	"context"
	"errors"
	"strings"
	"time"

	commonauth "photo-album/internal/common/auth"
	commonresponse "photo-album/internal/common/response"
	"photo-album/internal/svc"
	"photo-album/internal/types"
	"photo-album/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReviewPictureLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReviewPictureLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewPictureLogic {
	return &ReviewPictureLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReviewPictureLogic) ReviewPicture(req *types.ReviewPictureRequest, authorization string) (*types.PictureResponse, error) {
	if req == nil {
		return nil, commonresponse.BadRequest("请求体不能为空")
	}

	pictureID, err := parseRequiredSnowflakeID(req.Id, "id")
	if err != nil {
		return nil, err
	}
	if err := validateReviewDecision(req.ReviewStatus, req.ReviewMessage); err != nil {
		return nil, err
	}

	loginUser, err := commonauth.LoadRequiredLoginUser(l.ctx, l.svcCtx, authorization)
	if err != nil {
		return nil, err
	}

	pictureInfo, err := l.svcCtx.PicturesModel.FindOneActive(l.ctx, pictureID)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, commonresponse.NotFound("图片不存在")
		}
		return nil, commonresponse.InternalServerError("查询图片失败")
	}

	now := time.Now()
	pictureInfo.ReviewStatus = req.ReviewStatus
	pictureInfo.ReviewMessage = optionalString(strings.TrimSpace(req.ReviewMessage))
	pictureInfo.ReviewerId = optionalInt64(loginUser.Id)
	pictureInfo.ReviewTime = optionalTime(now)
	pictureInfo.UpdateTime = now

	if err := l.svcCtx.PicturesModel.Update(l.ctx, pictureInfo); err != nil {
		return nil, commonresponse.InternalServerError("审核图片失败")
	}

	userMap, err := loadUserDetailMap(l.ctx, l.svcCtx, []int64{pictureInfo.UserId})
	if err != nil {
		return nil, err
	}

	return buildPictureResponseWithUser(pictureInfo, userMap[pictureInfo.UserId], types.CompressPictureType{})
}

func validateReviewDecision(reviewStatus int64, reviewMessage string) error {
	switch reviewStatus {
	case reviewStatusPass:
		return nil
	case reviewStatusReject:
		if strings.TrimSpace(reviewMessage) == "" {
			return commonresponse.BadRequest("reviewMessage 不能为空")
		}
		return nil
	default:
		return commonresponse.BadRequest("reviewStatus 只能是 1 或 2")
	}
}
