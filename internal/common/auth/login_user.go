package auth

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	commonresponse "photo-album/internal/common/response"
	"photo-album/internal/svc"
	"photo-album/model"

	"github.com/golang-jwt/jwt/v4"
)

func ExtractUserIDFromBearerToken(authorization, secret string) (int64, error) {
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

func LoadRequiredLoginUser(ctx context.Context, svcCtx *svc.ServiceContext, authorization string) (*model.User, error) {
	if loginUser, ok := LoginUserFromContext(ctx); ok {
		return loginUser, nil
	}

	userID, err := ExtractUserIDFromBearerToken(authorization, svcCtx.Config.Auth.AccessSecret)
	if err != nil {
		return nil, commonresponse.Unauthorized("请先登录")
	}

	loginUser, err := svcCtx.UserModel.FindOneActive(ctx, userID)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, commonresponse.Unauthorized("登录用户不存在")
		}
		return nil, commonresponse.InternalServerError("查询登录用户失败")
	}

	return loginUser, nil
}

func LoadRequiredAdminUser(ctx context.Context, svcCtx *svc.ServiceContext, authorization string) (*model.User, error) {
	loginUser, err := LoadRequiredLoginUser(ctx, svcCtx, authorization)
	if err != nil {
		return nil, err
	}
	if loginUser.UserRole != "admin" {
		return nil, commonresponse.Forbidden("仅管理员可访问")
	}

	return loginUser, nil
}

func LoadOptionalLoginUser(ctx context.Context, svcCtx *svc.ServiceContext, authorization string) (*model.User, error) {
	if strings.TrimSpace(authorization) == "" {
		return nil, nil
	}

	return LoadRequiredLoginUser(ctx, svcCtx, authorization)
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
