package picture

import (
	"context"
	"errors"
	"strings"
	"time"

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

func (l *ReviewPictureLogic) ReviewPicture(req *types.PictureReviewRequest, authorization string) (*types.PictureResponse, error) {
	if req == nil || req.Id <= 0 {
		return nil, commonresponse.BadRequest("id 必须是正整数")
	}
	if err := validateReviewDecision(req.ReviewStatus, req.ReviewMessage); err != nil {
		return nil, err
	}

	loginUser, err := loadRequiredAdmin(l.ctx, l.svcCtx, authorization)
	if err != nil {
		return nil, err
	}

	pictureInfo, err := l.svcCtx.PicturesModel.FindOneActive(l.ctx, req.Id.Int64())
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

	return buildPictureResponse(pictureInfo), nil
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
