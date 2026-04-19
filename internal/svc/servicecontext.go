// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"photo-album/internal/config"
	"photo-album/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config        config.Config
	UserModel     model.UserModel
	PicturesModel model.PicturesModel
}

func NewServiceContext(c config.Config) *ServiceContext {

	conn := sqlx.NewMysql(c.Mysql.DataSource)

	return &ServiceContext{
		Config:        c,
		UserModel:     model.NewUserModel(conn, c.CacheRedis),
		PicturesModel: model.NewPicturesModel(conn, c.CacheRedis),
	}
}
