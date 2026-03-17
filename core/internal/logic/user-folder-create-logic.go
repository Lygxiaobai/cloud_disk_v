// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"cloud_disk/core/internal/helper"
	"cloud_disk/core/internal/logger"
	"cloud_disk/core/internal/models"
	"context"
	"errors"
	"fmt"

	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserFolderCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserFolderCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserFolderCreateLogic {
	return &UserFolderCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserFolderCreateLogic) UserFolderCreate(req *types.UserFolderCreateRequest, userIdentity string) (resp *types.UserFolderCreateResponse, err error) {
	// 从 context 中获取 TraceID
	traceID, _ := l.ctx.Value("trace_id").(string)
	ctx := context.WithValue(l.ctx, "method", "PUT")
	ctx = context.WithValue(ctx, "path", "/user/folder/create")
	ctx = context.WithValue(ctx, "user_identity", userIdentity)
	ctx = context.WithValue(ctx, "trace_id", traceID)

	//1.先查看当前层级有无这个同名的文件
	count, err := l.svcCtx.Engine.Where("name = ? AND parent_id = ?", req.Name, req.ParentId).Count(&models.UserRepository{})
	//有 返回错误信息
	if err != nil {
		logger.LogError(ctx, "查询文件夹失败", err, map[string]interface{}{
			"folder_name": req.Name,
			"parent_id":   req.ParentId,
		})
		return nil, err
	}
	if count > 0 {
		err = errors.New(fmt.Sprintf("%s already exists", req.Name))
		logger.LogError(ctx, "创建文件夹失败", err, map[string]interface{}{
			"folder_name": req.Name,
			"parent_id":   req.ParentId,
			"reason":      "文件夹已存在",
		})
		return nil, err
	}
	up := models.UserRepository{
		Identity:     helper.UUID(),
		Name:         req.Name,
		ParentId:     req.ParentId,
		UserIdentity: userIdentity,
		IsDir:        1,
	}
	//没有 则创建
	_, err = l.svcCtx.Engine.Insert(&up)
	if err != nil {
		logger.LogError(ctx, "插入文件夹失败", err, map[string]interface{}{
			"folder_name": req.Name,
			"parent_id":   req.ParentId,
		})
		return nil, err
	}
	resp = &types.UserFolderCreateResponse{
		Identity: up.Identity,
	}
	//返回创建后生成的Identity
	return
}
