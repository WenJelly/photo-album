package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ PicturesModel = (*customPicturesModel)(nil)

type (
	PictureStats struct {
		Total         int64 `db:"total"`
		ApprovedCount int64 `db:"approvedCount"`
		PendingCount  int64 `db:"pendingCount"`
		RejectedCount int64 `db:"rejectedCount"`
	}

	PicturesModel interface {
		picturesModel
		FindOneActive(ctx context.Context, id int64) (*Pictures, error)
		IncrementViewCount(ctx context.Context, id int64) error
		CountByWhere(ctx context.Context, whereSQL string, args ...any) (int64, error)
		FindByWhere(ctx context.Context, whereSQL, orderSQL string, limit, offset int64, args ...any) ([]*Pictures, error)
		CountStatsByUser(ctx context.Context, userID int64) (*PictureStats, error)
	}

	customPicturesModel struct {
		*defaultPicturesModel
	}
)

func NewPicturesModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) PicturesModel {
	return &customPicturesModel{
		defaultPicturesModel: newPicturesModel(conn, c, opts...),
	}
}

func (m *customPicturesModel) FindOneActive(ctx context.Context, id int64) (*Pictures, error) {
	var resp Pictures
	query := fmt.Sprintf("select %s from %s where `id` = ? and `isDelete` = 0 limit 1", picturesRows, m.table)
	if err := m.QueryRowNoCacheCtx(ctx, &resp, query, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &resp, nil
}

func (m *customPicturesModel) IncrementViewCount(ctx context.Context, id int64) error {
	picturesIdKey := fmt.Sprintf("%s%v", cachePicturesIdPrefix, id)
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		query := fmt.Sprintf("update %s set `viewCount` = `viewCount` + 1 where `id` = ? and `isDelete` = 0", m.table)
		return conn.ExecCtx(ctx, query, id)
	}, picturesIdKey)
	return err
}

func (m *customPicturesModel) CountByWhere(ctx context.Context, whereSQL string, args ...any) (int64, error) {
	var resp struct {
		Count int64 `db:"count"`
	}

	query := fmt.Sprintf("select count(1) as count from %s %s", m.table, strings.TrimSpace(whereSQL))
	if err := m.QueryRowNoCacheCtx(ctx, &resp, query, args...); err != nil {
		if errors.Is(err, ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}

	return resp.Count, nil
}

func (m *customPicturesModel) FindByWhere(ctx context.Context, whereSQL, orderSQL string, limit, offset int64, args ...any) ([]*Pictures, error) {
	orderSQL = strings.TrimSpace(orderSQL)
	if orderSQL == "" {
		orderSQL = "`id` desc"
	}

	query := fmt.Sprintf("select %s from %s %s order by %s limit ? offset ?",
		picturesRows, m.table, strings.TrimSpace(whereSQL), orderSQL)

	finalArgs := append([]any{}, args...)
	finalArgs = append(finalArgs, limit, offset)

	var resp []*Pictures
	if err := m.QueryRowsNoCacheCtx(ctx, &resp, query, finalArgs...); err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return resp, nil
}

func (m *customPicturesModel) CountStatsByUser(ctx context.Context, userID int64) (*PictureStats, error) {
	if userID <= 0 {
		return &PictureStats{}, nil
	}

	var resp PictureStats
	query := fmt.Sprintf(`select
		count(1) as total,
		coalesce(sum(case when reviewStatus = 1 then 1 else 0 end), 0) as approvedCount,
		coalesce(sum(case when reviewStatus = 0 then 1 else 0 end), 0) as pendingCount,
		coalesce(sum(case when reviewStatus = 2 then 1 else 0 end), 0) as rejectedCount
	from %s
	where isDelete = 0 and userId = ?`, m.table)

	if err := m.QueryRowNoCacheCtx(ctx, &resp, query, userID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return &PictureStats{}, nil
		}
		return nil, err
	}

	return &resp, nil
}
