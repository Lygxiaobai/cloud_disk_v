// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"cloud_disk/core/internal/define"
	"cloud_disk/core/internal/logger"
	"cloud_disk/core/internal/models"
	"context"
	"errors"

	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserFileListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserFileListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserFileListLogic {
	return &UserFileListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserFileListLogic) UserFileList(req *types.UserFileListRequest, userIdentity string) (resp *types.UserFileListResponse, err error) {
	// 从 context 中获取 TraceID
	traceID, _ := l.ctx.Value("trace_id").(string)
	ctx := context.WithValue(l.ctx, "method", "GET")
	ctx = context.WithValue(ctx, "path", "/user/file/list")
	ctx = context.WithValue(ctx, "user_identity", userIdentity)
	ctx = context.WithValue(ctx, "trace_id", traceID)

	list := []*types.UserFile{}
	size := req.Size
	if size == 0 {
		size = define.PageSize
	}
	page := req.Page
	if page == 0 {
		page = define.Page
	}
	offset := (page - 1) * size
	parentID := req.Id

	// 优先按目录唯一凭证查询，这样前端进入目录时不需要先知道数据库主键 id。
	if req.Identity != "" {
		folder := new(models.UserRepository)
		has, err := l.svcCtx.Engine.
			Where("identity = ? AND user_identity = ? AND is_dir = 1", req.Identity, userIdentity).
			Get(folder)
		if err != nil {
			logger.LogError(ctx, "查询文件夹失败", err, map[string]interface{}{
				"folder_identity": req.Identity,
			})
			return nil, err
		}
		if !has {
			err = errors.New("folder not found")
			logger.LogError(ctx, "查询文件列表失败", err, map[string]interface{}{
				"folder_identity": req.Identity,
				"reason":          "文件夹不存在",
			})
			return nil, err
		}
		parentID = int64(folder.Id)
	}

	l.svcCtx.Engine.ShowSQL(true)
	// 1. 查询当前目录下的所有直接子内容，包括文件夹和文件。
	err = l.svcCtx.Engine.Table("user_repository").Where("user_identity=? AND parent_id =?", userIdentity, parentID).
		Select("user_repository.id,user_repository.identity,user_repository.repository_identity,user_repository.name,user_repository.ext,user_repository.is_dir,repository_pool.size,repository_pool.path").
		Join("LEFT", "repository_pool", "user_repository.repository_identity = repository_pool.identity").
		Limit(size, offset).
		Where("user_repository.deleted_at IS NULL").
		Find(&list)
	if err != nil {
		logger.LogError(ctx, "查询文件列表失败", err, map[string]interface{}{
			"parent_id": parentID,
			"page":      page,
			"size":      size,
		})
		return nil, err
	}
	count, err := l.svcCtx.Engine.Where("user_identity=? AND parent_id =?", userIdentity, parentID).Count(&models.UserRepository{})
	if err != nil {
		logger.LogError(ctx, "统计文件数量失败", err, map[string]interface{}{
			"parent_id": parentID,
		})
		return nil, err
	}
	// 2. 返回当前目录内容列表和总数。
	resp = &types.UserFileListResponse{
		list,
		count,
	}
	return
}
