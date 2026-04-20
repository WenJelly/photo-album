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
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type RegisterUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterUserLogic {
	return &RegisterUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterUserLogic) RegisterUser(req *types.RegisterRequest) (resp *types.RegisterResponse, err error) {
	if req.UserPassword != req.UserCheckPassword {
		return nil, commonresponse.BadRequest("两次输入的密码不一致")
	}

	_, err = l.svcCtx.UserModel.FindOneByUserEmail(l.ctx, req.UserEmail)
	if err == nil {
		return nil, commonresponse.Conflict("邮箱已注册")
	}
	if !errors.Is(err, model.ErrNotFound) {
		return nil, commonresponse.InternalServerError("数据库查询异常")
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(req.UserPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, commonresponse.InternalServerError("密码加密失败")
	}

	now := time.Now()
	newUser := &model.User{
		UserEmail:    req.UserEmail,
		UserPassword: string(hashPassword),
		UserName:     "用户",
		UserRole:     "user",
		EditTime:     now,
		CreateTime:   now,
		UpdateTime:   now,
	}

	res, err := l.svcCtx.UserModel.Insert(l.ctx, newUser)
	if err != nil {
		return nil, commonresponse.InternalServerError("注册失败，请稍后重试")
	}

	newUserID, _ := res.LastInsertId()
	return &types.RegisterResponse{
		Id: types.NewSnowflakeID(newUserID),
	}, nil
}
