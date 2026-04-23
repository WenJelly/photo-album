// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"
	"errors"
	commonresponse "photo-album/internal/common/response"
	"photo-album/internal/svc"
	"photo-album/internal/types"
	"photo-album/model"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type LoginUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginUserLogic {
	return &LoginUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginUserLogic) LoginUser(req *types.LoginRequest) (resp *types.LoginResponse, err error) {
	userInfo, err := l.svcCtx.UserModel.FindOneByUserEmail(l.ctx, req.UserEmail)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, commonresponse.NotFound("账号不存在")
		}

		return nil, commonresponse.InternalServerError("数据库查询异常")
	}
	if userInfo.IsDelete != 0 {
		return nil, commonresponse.NotFound("账号不存在")
	}

	if err = bcrypt.CompareHashAndPassword([]byte(userInfo.UserPassword), []byte(req.UserPassword)); err != nil {
		return nil, commonresponse.Unauthorized("密码错误")
	}

	now := time.Now().Unix()
	accessExpire := l.svcCtx.Config.Auth.AccessExpire
	jwtToken, err := l.getJwtToken(l.svcCtx.Config.Auth.AccessSecret, now, accessExpire, userInfo.Id)
	if err != nil {
		return nil, commonresponse.InternalServerError("生成 Token 失败")
	}

	return &types.LoginResponse{
		Token:       jwtToken,
		Id:          formatUserID(userInfo.Id),
		UserEmail:   userInfo.UserEmail,
		UserName:    userInfo.UserName,
		UserAvatar:  userInfo.UserAvatar,
		UserProfile: userInfo.UserProfile,
		UserRole:    userInfo.UserRole,
		CreateTime:  userInfo.CreateTime.Format("2006-01-02 15:04:05"),
		UpdateTime:  userInfo.UpdateTime.Format("2006-01-02 15:04:05"),
	}, nil
}

func (l *LoginUserLogic) getJwtToken(secretKey string, iat, seconds, userID int64) (string, error) {
	claims := make(jwt.MapClaims)
	claims["exp"] = iat + seconds
	claims["iat"] = iat
	claims["userId"] = strconv.FormatInt(userID, 10)

	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims

	return token.SignedString([]byte(secretKey))
}
