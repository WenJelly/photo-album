package picture

import (
	"context"

	commonauth "photo-album/internal/common/auth"
	"photo-album/internal/svc"
	"photo-album/model"
)

func loadRequiredLoginUser(ctx context.Context, svcCtx *svc.ServiceContext, authorization string) (*model.User, error) {
	return commonauth.LoadRequiredLoginUser(ctx, svcCtx, authorization)
}

func loadOptionalLoginUser(ctx context.Context, svcCtx *svc.ServiceContext, authorization string) (*model.User, error) {
	return commonauth.LoadOptionalLoginUser(ctx, svcCtx, authorization)
}
