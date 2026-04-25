package user

import (
	"context"
	"errors"
	"strings"

	commonresponse "photo-album/internal/common/response"
	"photo-album/internal/svc"
	"photo-album/internal/types"
	"photo-album/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type DetailUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDetailUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DetailUserLogic {
	return &DetailUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DetailUserLogic) DetailUser(req *types.DetailUserRequest, authorization string) (*types.DetailUserResponse, error) {
	loginUser, err := loadRequiredLoginUser(l.ctx, l.svcCtx, authorization)
	if err != nil {
		return nil, err
	}

	if req == nil {
		return loadActiveUserDetail(l.ctx, l.svcCtx, loginUser.Id)
	}

	if strings.TrimSpace(req.Id) != "" {
		userID, err := parseRequiredUserID(req.Id)
		if err != nil {
			return nil, err
		}
		return loadActiveUserDetail(l.ctx, l.svcCtx, userID)
	}

	if userEmail := strings.TrimSpace(req.UserEmail); userEmail != "" {
		userInfo, err := l.svcCtx.UserModel.FindOneByUserEmail(l.ctx, userEmail)
		if err != nil {
			if errors.Is(err, model.ErrNotFound) || (userInfo != nil && userInfo.IsDelete != 0) {
				return nil, commonresponse.NotFound("用户不存在")
			}
			return nil, commonresponse.InternalServerError("查询用户信息失败")
		}
		if userInfo.IsDelete != 0 {
			return nil, commonresponse.NotFound("用户不存在")
		}
		return loadActiveUserDetail(l.ctx, l.svcCtx, userInfo.Id)
	}

	return loadActiveUserDetail(l.ctx, l.svcCtx, loginUser.Id)
}
