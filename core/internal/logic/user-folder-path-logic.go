// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/models"
	"context"

	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserFolderPathLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserFolderPathLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserFolderPathLogic {
	return &UserFolderPathLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 从当前目录到根目录的路径 然后进行反转
func (l *UserFolderPathLogic) UserFolderPath(req *types.UserFolderPathRequest, userIdentity string) (resp *types.UserFolderPathResponse, err error) {
	list := []*types.FolderPathItem{}
	//identity = ""
	if req.Identity == "" {
		list = append(list, &types.FolderPathItem{
			Identity: "",
			Name:     "全部文件",
			Id:       0,
		})
		return &types.UserFolderPathResponse{
			List: list,
		}, nil
	}
	//identity 有值
	curr := models.UserRepository{}
	has, err := l.svcCtx.Engine.Where("identity=? and is_dir = 1 and user_identity = ?", req.Identity, userIdentity).Get(&curr)
	if err != nil {
		return nil, errors.New(l.ctx, "查询文件夹失败", err, map[string]interface{}{
			"folder_identity": req.Identity,
		})
	}
	if !has {
		return nil, errors.New(l.ctx, "查询文件夹路径失败", nil, map[string]interface{}{
			"folder_identity": req.Identity,
			"reason":          "不是文件夹",
		})
	}
	for {
		list = append(list, &types.FolderPathItem{
			Identity: curr.Identity,
			Name:     curr.Name,
			Id:       int64(curr.Id),
		})
		if curr.ParentId == 0 {
			break
		}
		parent := models.UserRepository{}
		has, err = l.svcCtx.Engine.Where("id=? AND user_identity = ? AND is_dir = 1", curr.ParentId, userIdentity).Get(&parent)
		if err != nil {
			return nil, errors.New(l.ctx, "查询父级文件夹失败", err, map[string]interface{}{
				"parent_id": curr.ParentId,
			})
		}
		if !has {
			return nil, errors.New(l.ctx, "查询文件夹路径失败", nil, map[string]interface{}{
				"parent_id": curr.ParentId,
				"reason":    "父级文件夹不存在",
			})
		}
		curr = parent

	}
	//反转 路径
	for left, right := 0, len(list)-1; left < right; left, right = left+1, right-1 {
		list[left], list[right] = list[right], list[left]
	}
	// 在最前面补一个虚拟根节点，前端面包屑展示更统一。
	list = append([]*types.FolderPathItem{
		{
			Id:       0,
			Identity: "",
			Name:     "全部文件",
		},
	}, list...)

	resp = &types.UserFolderPathResponse{
		List: list,
	}
	return
}
