package model

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserModel = (*customUserModel)(nil)

type (
	// UserModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserModel.
	UserModel interface {
		userModel
		FindOneActive(ctx context.Context, id int64) (*User, error)
		FindByIDs(ctx context.Context, ids []int64) ([]*User, error)
	}

	customUserModel struct {
		*defaultUserModel
	}
)

// NewUserModel returns a model for the database table.
func NewUserModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) UserModel {
	return &customUserModel{
		defaultUserModel: newUserModel(conn, c, opts...),
	}
}

func (m *customUserModel) FindOneActive(ctx context.Context, id int64) (*User, error) {
	var resp User
	query := fmt.Sprintf("select %s from %s where `id` = ? and `isDelete` = 0 limit 1", userRows, m.table)
	if err := m.QueryRowNoCacheCtx(ctx, &resp, query, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &resp, nil
}

func (m *customUserModel) FindByIDs(ctx context.Context, ids []int64) ([]*User, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	placeholders := make([]string, 0, len(ids))
	args := make([]any, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	if len(placeholders) == 0 {
		return nil, nil
	}

	query := fmt.Sprintf("select %s from %s where `isDelete` = 0 and `id` in (%s)",
		userRows, m.table, strings.Join(placeholders, ","))

	var users []*User
	if err := m.QueryRowsNoCacheCtx(ctx, &users, query, args...); err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return users, nil
}
