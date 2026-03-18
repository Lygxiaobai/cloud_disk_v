// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"cloud_disk/core/internal/logger"
	"context"

	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ShareFileDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewShareFileDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ShareFileDetailLogic {
	return &ShareFileDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ShareFileDetailLogic) ShareFileDetail(req *types.ShareFileDetailRequest) (resp *types.ShareFileDetailResponse, err error) {
	// 从 context 中获取 TraceID
	traceID, _ := l.ctx.Value("trace_id").(string)
	ctx := context.WithValue(l.ctx, "method", "GET")
	ctx = context.WithValue(ctx, "path", "/share/file/detail")
	ctx = context.WithValue(ctx, "trace_id", traceID)

	// 1. 判断是否是热门分享
	isHot := l.svcCtx.ShareCache.IsHotShare(ctx, req.Identity)

	// 2. 如果是热门分享，优先从 Redis 读取
	if isHot {
		resp, err = l.svcCtx.ShareCache.GetShareDetail(ctx, req.Identity)
		if err != nil {
			logger.LogError(ctx, "从Redis读取分享详情失败", err, map[string]interface{}{
				"share_identity": req.Identity,
			})
		} else if resp != nil {
			// 缓存命中，更新点击次数（异步，不影响响应速度）
			go func() {
				// 更新日榜点击数
				l.svcCtx.ShareCache.IncrDailyClick(context.Background(), req.Identity)
				// 更新数据库总点击数和更新时间
				l.svcCtx.Engine.Exec("update share_basic set click_num = click_num + 1, updated_at = NOW() where identity = ?", req.Identity)
			}()
			return resp, nil
		}
	}

	// 3. 缓存未命中或不是热门分享，查询数据库
	// 更新点击次数和更新时间
	_, err = l.svcCtx.Engine.Exec("update share_basic set click_num = click_num + 1, updated_at = NOW() where identity = ?", req.Identity)
	if err != nil {
		logger.LogError(ctx, "更新分享点击次数失败", err, map[string]interface{}{
			"share_identity": req.Identity,
		})
		return nil, err
	}

	// 更新日榜点击数（异步）
	go func() {
		l.svcCtx.ShareCache.IncrDailyClick(context.Background(), req.Identity)
	}()

	// 查询分享详情
	resp = &types.ShareFileDetailResponse{}
	_, err = l.svcCtx.Engine.Table("share_basic").
		Select("share_basic.repository_identity, user_repository.name, user_repository.ext,repository_pool.size,repository_pool.path").
		Join("LEFT", "repository_pool", "share_basic.repository_identity = repository_pool.identity").
		Join("LEFT", "user_repository", "share_basic.user_repository_identity = user_repository.identity").
		Where("share_basic.identity = ?", req.Identity).Get(resp)
	if err != nil {
		logger.LogError(ctx, "查询分享文件详情失败", err, map[string]interface{}{
			"share_identity": req.Identity,
		})
		return nil, err
	}

	// 4. 如果是热门分享，写入缓存
	if isHot {
		err = l.svcCtx.ShareCache.SetShareDetail(ctx, req.Identity, resp)
		if err != nil {
			logger.LogError(ctx, "写入分享详情到Redis失败", err, map[string]interface{}{
				"share_identity": req.Identity,
			})
			// 写入缓存失败不影响业务，继续返回
		}
	}

	return
}
