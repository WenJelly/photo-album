package auth

import (
	"context"

	"photo-album/model"
)

type loginUserContextKey struct{}

func WithLoginUser(ctx context.Context, user *model.User) context.Context {
	if ctx == nil || user == nil {
		return ctx
	}

	return context.WithValue(ctx, loginUserContextKey{}, user)
}

func LoginUserFromContext(ctx context.Context) (*model.User, bool) {
	if ctx == nil {
		return nil, false
	}

	user, ok := ctx.Value(loginUserContextKey{}).(*model.User)
	return user, ok && user != nil
}
