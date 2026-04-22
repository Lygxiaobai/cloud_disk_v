package logic

import (
	"cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/models"
	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
)

// UserFilePreviewLogic 负责文件预览。
// 它会根据文件类型决定是返回签名 URL，还是返回文本片段。
type UserFilePreviewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserFilePreviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserFilePreviewLogic {
	return &UserFilePreviewLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UserFilePreview 处理预览逻辑：
// 1. 检查文件归属
// 2. 读取 repository_pool
// 3. 判断预览类型
// 4. 图片/视频/PDF 返回签名 URL，文本返回片段
// 5. 顺手写入“最近文件”
func (l *UserFilePreviewLogic) UserFilePreview(req *types.UserFilePreviewRequest, userIdentity string) (*types.UserFilePreviewResponse, error) {
	sess := l.svcCtx.Engine.NewSession()
	defer sess.Close()

	file, err := getUserRepository(l.ctx, sess, userIdentity, req.Identity, false, false)
	if err != nil {
		return nil, err
	}
	if file.IsDir == 1 {
		return nil, errors.New(l.ctx, "folder does not support preview", nil, map[string]interface{}{
			"identity": req.Identity,
		})
	}

	repository := new(models.RepositoryPool)
	has, err := sess.Where("identity = ?", file.RepositoryIdentity).Get(repository)
	if err != nil {
		return nil, errors.New(l.ctx, "query repository pool failed", err, map[string]interface{}{
			"repository_identity": file.RepositoryIdentity,
		})
	}
	if !has {
		return nil, errors.New(l.ctx, "repository does not exist", nil, map[string]interface{}{
			"repository_identity": file.RepositoryIdentity,
		})
	}

	kind := previewKind(file.Ext)
	resp := &types.UserFilePreviewResponse{
		Kind: kind,
		Name: file.Name,
		Ext:  file.Ext,
		Size: repository.Size,
	}

	// 新数据优先用 object_key；
	// 历史数据可能只有 path，因此这里兼容恢复 objectKey。
	objectKey := repository.ObjectKey
	if objectKey == "" {
		objectKey = l.svcCtx.OSS.GuessObjectKey(repository.Path)
	}

	if objectKey == "" {
		// 如果连 objectKey 都恢复不出来，就只能退回旧 path 方案。
		if repository.Path == "" {
			return nil, errors.New(l.ctx, "preview resource path is empty", nil, map[string]interface{}{
				"repository_identity": repository.Identity,
			})
		}
		if kind == "text" {
			// 文本预览在没有 objectKey 时无法做 Range 读取，因此降级为下载。
			resp.Kind = "download"
		}
		resp.URL = repository.Path
		addRecentFile(context.Background(), l.svcCtx.RDB, userIdentity, file.Identity)
		return resp, nil
	}

	switch kind {
	case "text":
		// 文本文件只展示前一小段内容，避免大文件预览拖垮接口。
		body, truncated, err := l.svcCtx.OSS.ReadObjectRange(l.ctx, objectKey, 4096)
		if err != nil {
			return nil, errors.New(l.ctx, "read preview text failed", err, map[string]interface{}{
				"object_key": objectKey,
			})
		}
		resp.Text = strings.ToValidUTF8(string(body), "")
		resp.Truncated = truncated
	default:
		// 图片、视频、音频、PDF 等都走签名 URL，浏览器直接渲染。
		url, err := l.svcCtx.OSS.SignGetObjectURL(l.ctx, objectKey, l.svcCtx.OSS.PreviewExpires(), file.Name)
		if err != nil {
			// 如果签名失败，至少回退成普通对象地址，方便排查或兜底访问。
			resp.URL = l.svcCtx.OSS.BuildObjectURL(objectKey)
			addRecentFile(context.Background(), l.svcCtx.RDB, userIdentity, file.Identity)
			return resp, nil
		}
		resp.URL = url
	}

	addRecentFile(context.Background(), l.svcCtx.RDB, userIdentity, file.Identity)
	return resp, nil
}
