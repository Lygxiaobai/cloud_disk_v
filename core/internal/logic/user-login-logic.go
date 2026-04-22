// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/helper"
	"cloud_disk/core/internal/models"
	"context"
	"fmt"
	"time"

	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserLoginLogic {
	return &UserLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserLoginLogic) UserLogin(req *types.LoginRequest) (resp *types.LoginResponse, err error) {
	lockKey := "loginlock:" + req.Name
	if locked, _ := l.svcCtx.RDB.Get(l.ctx, lockKey).Result(); locked == "1" {
		ttl, _ := l.svcCtx.RDB.TTL(l.ctx, lockKey).Result()
		minutes := int(ttl.Minutes())
		if minutes <= 0 {
			minutes = 1
		}
		return nil, errors.LoginLocked(l.ctx, fmt.Sprintf("账户已锁定，请 %d 分钟后重试", minutes), nil, map[string]interface{}{
			"username": req.Name,
		})
	}

	user := new(models.UserBasic)
	has, err := l.svcCtx.Engine.Where("name = ?", req.Name).Get(user)
	if err != nil {
		return nil, errors.Internal(l.ctx, "数据库查询失败", err, map[string]interface{}{
			"username": req.Name,
		})
	}
	if !has {
		l.recordLoginFailure(req.Name)
		return nil, errors.AuthFailed(l.ctx, "用户名或密码错误", nil, map[string]interface{}{
			"username": req.Name,
			"reason":   "user_not_found",
		})
	}

	if err := helper.CheckPassword(user.Password, req.Password); err != nil {
		l.recordLoginFailure(req.Name)
		return nil, errors.AuthFailed(l.ctx, "用户名或密码错误", nil, map[string]interface{}{
			"username": req.Name,
			"reason":   "password_mismatch",
		})
	}

	l.svcCtx.RDB.Del(l.ctx, "loginfail:"+req.Name)

	jwtCfg := l.svcCtx.Config.JWT
	token, err := helper.GenerateToken(user.Id, user.Identity, user.Name, user.Role,
		jwtCfg.AccessSecret, jwtCfg.AccessExpire)
	if err != nil {
		return nil, errors.Internal(l.ctx, "生成Token失败", err, map[string]interface{}{
			"user_id":   user.Id,
			"user_name": user.Name,
		})
	}
	refreshToken, err := helper.GenerateToken(user.Id, user.Identity, user.Name, user.Role,
		jwtCfg.RefreshSecret, jwtCfg.RefreshExpire)
	if err != nil {
		return nil, errors.Internal(l.ctx, "生成RefreshToken失败", err, map[string]interface{}{
			"user_id":   user.Id,
			"user_name": user.Name,
		})
	}

	resp = &types.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
	}
	return resp, nil
}

func (l *UserLoginLogic) recordLoginFailure(name string) {
	rlCfg := l.svcCtx.Config.RateLimit
	if rlCfg.LoginLockThreshold <= 0 {
		return
	}

	failKey := "loginfail:" + name
	fails, err := l.svcCtx.RDB.Incr(l.ctx, failKey).Result()
	if err != nil {
		logx.WithContext(l.ctx).Errorf("incr login fail counter failed: %v", err)
		return
	}
	if fails == 1 {
		l.svcCtx.RDB.Expire(l.ctx, failKey, 15*time.Minute)
	}
	if fails >= int64(rlCfg.LoginLockThreshold) {
		lockMinutes := rlCfg.LoginLockMinutes
		if lockMinutes <= 0 {
			lockMinutes = 15
		}
		l.svcCtx.RDB.Set(l.ctx, "loginlock:"+name, "1", time.Duration(lockMinutes)*time.Minute)
		l.svcCtx.RDB.Del(l.ctx, failKey)
	}
}
