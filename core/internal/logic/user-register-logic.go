package logic

import (
	"cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/helper"
	"cloud_disk/core/internal/models"
	"context"

	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserRegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserRegisterLogic {
	return &UserRegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserRegisterLogic) UserRegister(req *types.UserRegisterRequest) (resp *types.UserRegisterResponse, err error) {
	redisKey := "code:" + req.Email
	code, err := l.svcCtx.RDB.Get(l.ctx, redisKey).Result()
	if err != nil || code != req.Code {
		return nil, errors.VerificationFailed(l.ctx, "验证码验证失败", err, map[string]interface{}{
			"email": req.Email,
		})
	}

	sess := l.svcCtx.Engine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return nil, errors.Internal(l.ctx, "启动注册事务失败", err, nil)
	}

	committed := false
	defer func() {
		if committed {
			return
		}
		if rbErr := sess.Rollback(); rbErr != nil {
			logx.WithContext(l.ctx).Errorf("rollback register failed: %v", rbErr)
		}
	}()

	count := int64(0)
	var emailRow struct {
		Count int64 `xorm:"'count'"`
	}
	has, err := sess.SQL("SELECT COUNT(1) AS count FROM user_basic WHERE email = ? FOR UPDATE", req.Email).Get(&emailRow)
	if err != nil {
		return nil, errors.Internal(l.ctx, "查询邮箱失败", err, map[string]interface{}{
			"email": req.Email,
		})
	}
	if has {
		count = emailRow.Count
	}
	if count > 0 {
		return nil, errors.Conflict(l.ctx, "邮箱已被注册", nil, map[string]interface{}{
			"email": req.Email,
		})
	}

	var nameRow struct {
		Count int64 `xorm:"'count'"`
	}
	has, err = sess.SQL("SELECT COUNT(1) AS count FROM user_basic WHERE name = ? FOR UPDATE", req.Name).Get(&nameRow)
	if err != nil {
		return nil, errors.Internal(l.ctx, "查询用户名失败", err, map[string]interface{}{
			"username": req.Name,
		})
	}
	count = 0
	if has {
		count = nameRow.Count
	}
	if count > 0 {
		return nil, errors.Conflict(l.ctx, "用户名已被注册", nil, map[string]interface{}{
			"username": req.Name,
		})
	}

	hashed, err := helper.HashPassword(req.Password)
	if err != nil {
		return nil, errors.Internal(l.ctx, "密码加密失败", err, nil)
	}

	user := models.UserBasic{
		Identity: helper.UUID(),
		Name:     req.Name,
		Email:    req.Email,
		Password: hashed,
	}
	_, err = sess.InsertOne(&user)
	if err != nil {
		return nil, errors.Internal(l.ctx, "插入用户失败", err, map[string]interface{}{
			"username": req.Name,
			"email":    req.Email,
		})
	}

	if err := sess.Commit(); err != nil {
		return nil, errors.Internal(l.ctx, "提交注册事务失败", err, nil)
	}
	committed = true
	return
}
