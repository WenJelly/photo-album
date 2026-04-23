package handler

import (
	"net/http"

	admin "photo-album/internal/handler/admin"
	picture "photo-album/internal/handler/picture"
	user "photo-album/internal/handler/user"
	logicpicture "photo-album/internal/logic/picture"
	middleware "photo-album/internal/middleware"
	"photo-album/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	adminCheckMiddleware := middleware.NewAdminCheckMiddleware(serverCtx)

	server.AddRoutes(
		rest.WithMiddleware(adminCheckMiddleware.Handle,
			[]rest.Route{
				{
					Method:  http.MethodPost,
					Path:    "/picture/list",
					Handler: picture.GetAdminPictureListHandler(serverCtx),
				},
				{
					Method:  http.MethodPost,
					Path:    "/user/update",
					Handler: admin.AdminUpdateUserHandler(serverCtx),
				},
			}...,
		),
		rest.WithJwt(serverCtx.Config.Auth.AccessSecret),
		rest.WithPrefix("/api/admin"),
	)

	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodPost,
				Path:    "/picture/list",
				Handler: picture.GetPictureListHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/user/login",
				Handler: user.LoginUserHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/user/register",
				Handler: user.RegisterUserHandler(serverCtx),
			},
		},
		rest.WithPrefix("/api"),
	)

	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodPost,
				Path:    "/picture/delete",
				Handler: picture.DeletePictureHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/picture/upload/url",
				Handler: picture.UploadPictureByUrlHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/picture/vo",
				Handler: picture.GetPictureVOHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/user/get/detail",
				Handler: user.DetailUserHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/user/update",
				Handler: user.UpdateUserHandler(serverCtx),
			},
		},
		rest.WithJwt(serverCtx.Config.Auth.AccessSecret),
		rest.WithPrefix("/api"),
	)

	server.AddRoutes(
		rest.WithMiddleware(adminCheckMiddleware.Handle,
			rest.Route{
				Method:  http.MethodPost,
				Path:    "/picture/review",
				Handler: picture.ReviewPictureHandler(serverCtx),
			},
		),
		rest.WithJwt(serverCtx.Config.Auth.AccessSecret),
		rest.WithPrefix("/api"),
	)

	server.AddRoute(
		rest.Route{
			Method:  http.MethodPost,
			Path:    "/picture/upload",
			Handler: picture.UploadPictureHandler(serverCtx),
		},
		rest.WithJwt(serverCtx.Config.Auth.AccessSecret),
		rest.WithPrefix("/api"),
		rest.WithMaxBytes(logicpicture.MaxFileUploadSize),
	)
}
