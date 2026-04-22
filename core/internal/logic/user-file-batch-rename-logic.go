package logic

import (
	"cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/models"
	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserFileBatchRenameLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserFileBatchRenameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserFileBatchRenameLogic {
	return &UserFileBatchRenameLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserFileBatchRenameLogic) UserFileBatchRename(req *types.UserFileBatchRenameRequest, userIdentity string) (*types.UserFileBatchRenameResponse, error) {
	identities := uniqueStrings(req.Identities)
	if len(identities) == 0 {
		return &types.UserFileBatchRenameResponse{List: []*types.UserFileBatchRenameItem{}}, nil
	}

	startIndex := req.StartIndex
	if startIndex == 0 {
		startIndex = 1
	}
	step := req.Step
	if step == 0 {
		step = 1
	}
	keepExt := true
	if req.KeepExt != nil {
		keepExt = *req.KeepExt
	}

	sess := l.svcCtx.Engine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return nil, errors.New(l.ctx, "start batch rename transaction failed", err, nil)
	}
	committed := false
	defer func() {
		if committed {
			return
		}
		if rbErr := sess.Rollback(); rbErr != nil {
			logx.WithContext(l.ctx).Errorf("rollback batch rename failed: %v", rbErr)
		}
	}()

	files, err := getUserRepositoriesByIdentities(l.ctx, sess, userIdentity, identities, false, true)
	if err != nil {
		return nil, err
	}

	items := make([]*types.UserFileBatchRenameItem, 0, len(identities))
	nextNamesByIdentity := make(map[string]string, len(identities))
	namesByParent := make(map[int64][]string)
	seenByParent := make(map[int64]map[string]string)

	for index, identity := range identities {
		file, ok := files[identity]
		if !ok {
			return nil, errors.New(l.ctx, "file does not exist", nil, map[string]interface{}{
				"identity": identity,
			})
		}

		nextName := buildBatchRenameName(file, req, keepExt, startIndex+index*step)
		nextNamesByIdentity[file.Identity] = nextName
		items = append(items, &types.UserFileBatchRenameItem{
			Identity: file.Identity,
			OldName:  file.Name,
			NewName:  nextName,
		})

		if file.Name == nextName {
			continue
		}
		if _, ok := seenByParent[file.ParentId]; !ok {
			seenByParent[file.ParentId] = make(map[string]string)
		}
		if conflictIdentity, exists := seenByParent[file.ParentId][nextName]; exists && conflictIdentity != file.Identity {
			return nil, errors.New(l.ctx, "same name already exists in current folder", nil, map[string]interface{}{
				"parent_id": file.ParentId,
				"name":      nextName,
			})
		}
		seenByParent[file.ParentId][nextName] = file.Identity
		namesByParent[file.ParentId] = append(namesByParent[file.ParentId], nextName)
	}

	for parentID, names := range namesByParent {
		existing, loadErr := loadExistingNames(l.ctx, sess, userIdentity, parentID, names, false, true)
		if loadErr != nil {
			return nil, loadErr
		}
		for _, identity := range identities {
			file, ok := files[identity]
			if !ok || file.ParentId != parentID {
				continue
			}
			nextName := nextNamesByIdentity[file.Identity]
			if nextName == file.Name {
				continue
			}
			for _, row := range existing[nextName] {
				if row.Identity != file.Identity {
					return nil, errors.New(l.ctx, "same name already exists in current folder", nil, map[string]interface{}{
						"parent_id": parentID,
						"name":      nextName,
					})
				}
			}
		}
	}

	caseSQL := "name = CASE identity"
	args := make([]interface{}, 0, len(identities)*3+1)
	updateIdentities := make([]string, 0, len(identities))
	for _, identity := range identities {
		file := files[identity]
		nextName := nextNamesByIdentity[file.Identity]
		if nextName == file.Name {
			continue
		}
		caseSQL += " WHEN ? THEN ?"
		args = append(args, file.Identity, nextName)
		updateIdentities = append(updateIdentities, file.Identity)
	}
	caseSQL += " ELSE name END"

	if len(updateIdentities) > 0 {
		sql := "UPDATE user_repository SET " + caseSQL + " WHERE user_identity = ? AND deleted_at IS NULL AND identity IN (" + placeholders(len(updateIdentities)) + ")"
		args = append(args, userIdentity)
		for _, identity := range updateIdentities {
			args = append(args, identity)
		}
		params := append([]interface{}{sql}, args...)
		if _, err := sess.Exec(params...); err != nil {
			return nil, errors.New(l.ctx, "batch rename file failed", err, map[string]interface{}{
				"identities": updateIdentities,
			})
		}
	}

	if err := sess.Commit(); err != nil {
		return nil, errors.New(l.ctx, "commit batch rename failed", err, nil)
	}
	committed = true

	return &types.UserFileBatchRenameResponse{List: items}, nil
}

func buildBatchRenameName(file *models.UserRepository, req *types.UserFileBatchRenameRequest, keepExt bool, sequence int) string {
	name := file.Name
	ext := ""
	baseName := name

	if file.IsDir != 1 && keepExt {
		ext = path.Ext(name)
		baseName = strings.TrimSuffix(name, ext)
	}

	if req.FindText != "" {
		baseName = strings.ReplaceAll(baseName, req.FindText, req.ReplaceText)
	}
	baseName = strings.TrimSpace(req.Prefix) + baseName + strings.TrimSpace(req.Suffix)

	if req.ApplySequence {
		number := fmt.Sprintf("%d", sequence)
		if req.Padding > 0 {
			number = fmt.Sprintf("%0*d", req.Padding, sequence)
		}
		baseName += number
	}

	baseName = strings.TrimSpace(baseName)
	if baseName == "" {
		baseName = "unnamed"
	}
	if file.IsDir == 1 || !keepExt {
		return baseName
	}
	return baseName + ext
}
