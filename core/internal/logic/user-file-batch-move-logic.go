package logic

import (
	"cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/models"
	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserFileBatchMoveLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserFileBatchMoveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserFileBatchMoveLogic {
	return &UserFileBatchMoveLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserFileBatchMoveLogic) UserFileBatchMove(req *types.UserFileBatchMoveRequest, userIdentity string) (*types.UserFileBatchMoveResponse, error) {
	identities := uniqueStrings(req.Identities)
	if len(identities) == 0 {
		return &types.UserFileBatchMoveResponse{}, nil
	}

	sess := l.svcCtx.Engine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return nil, errors.New(l.ctx, "start batch move transaction failed", err, nil)
	}
	committed := false
	defer func() {
		if committed {
			return
		}
		if err := sess.Rollback(); err != nil {
			logx.WithContext(l.ctx).Errorf("rollback batch move failed: %v", err)
		}
	}()

	targetParentID := int64(0)
	targetFolderIdentity := req.ParentIdentity
	if req.ParentIdentity != "" {
		targetFolder, err := getUserRepository(l.ctx, sess, userIdentity, req.ParentIdentity, false, true)
		if err != nil {
			return nil, err
		}
		if targetFolder.IsDir != 1 {
			return nil, errors.New(l.ctx, "target is not a folder", nil, map[string]interface{}{
				"parent_identity": req.ParentIdentity,
			})
		}
		targetParentID = int64(targetFolder.Id)
		targetFolderIdentity = targetFolder.Identity
	}

	files, err := getUserRepositoriesByIdentities(l.ctx, sess, userIdentity, identities, false, true)
	if err != nil {
		return nil, err
	}
	for _, identity := range identities {
		if _, ok := files[identity]; !ok {
			return nil, errors.New(l.ctx, "file does not exist", nil, map[string]interface{}{
				"identity": identity,
			})
		}
	}

	targetNames := make([]string, 0, len(identities))
	batchNameSet := make(map[string]struct{}, len(identities))
	for _, identity := range identities {
		file := files[identity]
		if targetFolderIdentity != "" && file.Identity == targetFolderIdentity {
			return nil, errors.New(l.ctx, "cannot move folder into itself", nil, map[string]interface{}{
				"identity": identity,
			})
		}
		if file.IsDir == 1 && targetFolderIdentity != "" {
			subtree, subtreeErr := collectSubtreeIdentities(l.ctx, sess, userIdentity, []string{file.Identity}, false, true)
			if subtreeErr != nil {
				return nil, subtreeErr
			}
			for _, childIdentity := range subtree {
				if childIdentity == targetFolderIdentity {
					return nil, errors.New(l.ctx, "cannot move folder into its child folder", nil, map[string]interface{}{
						"identity": identity,
					})
				}
			}
		}

		if file.ParentId == targetParentID {
			continue
		}
		if _, exists := batchNameSet[file.Name]; exists {
			return nil, errors.New(l.ctx, "same name already exists in current folder", nil, map[string]interface{}{
				"parent_id": targetParentID,
				"name":      file.Name,
			})
		}
		batchNameSet[file.Name] = struct{}{}
		targetNames = append(targetNames, file.Name)
	}

	existingNames, err := loadExistingNames(l.ctx, sess, userIdentity, targetParentID, targetNames, false, true)
	if err != nil {
		return nil, err
	}
	for _, identity := range identities {
		file := files[identity]
		if file.ParentId == targetParentID {
			continue
		}
		for _, existing := range existingNames[file.Name] {
			if existing.Identity != file.Identity {
				return nil, errors.New(l.ctx, "same name already exists in current folder", nil, map[string]interface{}{
					"parent_id": targetParentID,
					"name":      file.Name,
				})
			}
		}
	}

	if _, err := sess.Where("user_identity = ? AND deleted_at IS NULL", userIdentity).
		In("identity", identities).
		Cols("parent_id").
		Update(&models.UserRepository{ParentId: targetParentID}); err != nil {
		return nil, errors.New(l.ctx, "move file failed", err, map[string]interface{}{
			"identities":      identities,
			"parent_identity": req.ParentIdentity,
		})
	}

	if err := sess.Commit(); err != nil {
		return nil, errors.New(l.ctx, "commit batch move failed", err, nil)
	}
	committed = true
	return &types.UserFileBatchMoveResponse{}, nil
}
