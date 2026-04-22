package logic

import (
	"cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/helper"
	"cloud_disk/core/internal/models"
	"context"
	"strings"

	"xorm.io/xorm"
)

// snapshotUserFileVersion 在逻辑文件切换到新物理版本前，先把旧版本快照下来。
// 这样后续查看版本历史时，可以看到每一次替换前的资源信息。
func snapshotUserFileVersion(ctx context.Context, sess *xorm.Session, file *models.UserRepository, repository *models.RepositoryPool, action string) error {
	if file == nil || repository == nil {
		return nil
	}

	version := &models.UserFileVersion{
		Identity:           helper.UUID(),
		UserIdentity:       file.UserIdentity,
		FileIdentity:       file.Identity,
		RepositoryIdentity: repository.Identity,
		Name:               file.Name,
		Ext:                file.Ext,
		Size:               repository.Size,
		Hash:               repository.Hash,
		Action:             action,
	}
	if _, err := sess.Insert(version); err != nil {
		return errors.New(ctx, "save file version snapshot failed", err, map[string]interface{}{
			"file_identity": file.Identity,
			"action":        action,
		})
	}
	return nil
}

// replaceUserFileRepository 用于“上传新版本”场景。
// 它会保留当前逻辑文件 identity，只替换背后的 repository_identity，
// 这样分享、收藏、目录结构都不需要跟着重建。
func replaceUserFileRepository(ctx context.Context, sess *xorm.Session, userIdentity string, targetFileIdentity string, repository *models.RepositoryPool) (*models.UserRepository, error) {
	file, err := getUserRepository(ctx, sess, userIdentity, targetFileIdentity, false, true)
	if err != nil {
		return nil, err
	}
	if file.IsDir == 1 {
		return nil, errors.New(ctx, "folder does not support version replace", nil, map[string]interface{}{
			"file_identity": targetFileIdentity,
		})
	}

	// 为了避免“文档变视频”这类体验混乱，当前版本先限制新旧扩展名一致。
	if strings.TrimSpace(file.Ext) != "" && strings.TrimSpace(repository.Ext) != "" && !strings.EqualFold(file.Ext, repository.Ext) {
		return nil, errors.New(ctx, "new version must keep the same file extension", nil, map[string]interface{}{
			"file_identity": targetFileIdentity,
			"old_ext":       file.Ext,
			"new_ext":       repository.Ext,
		})
	}

	currentRepository := new(models.RepositoryPool)
	has, err := sess.Where("identity = ?", file.RepositoryIdentity).Get(currentRepository)
	if err != nil {
		return nil, errors.New(ctx, "query current repository failed", err, map[string]interface{}{
			"file_identity": targetFileIdentity,
		})
	}
	if has && currentRepository.Identity != repository.Identity {
		if err := snapshotUserFileVersion(ctx, sess, file, currentRepository, "replace"); err != nil {
			return nil, err
		}
	}

	updater := &models.UserRepository{
		RepositoryIdentity: repository.Identity,
		Ext:                repository.Ext,
	}
	if _, err := sess.Where("identity = ? AND user_identity = ?", targetFileIdentity, userIdentity).Cols("repository_identity", "ext").Update(updater); err != nil {
		return nil, errors.New(ctx, "replace file repository failed", err, map[string]interface{}{
			"file_identity":       targetFileIdentity,
			"repository_identity": repository.Identity,
		})
	}

	file.RepositoryIdentity = repository.Identity
	file.Ext = repository.Ext
	return file, nil
}

// restoreUserFileVersion switches the logical file back to one historical repository version.
// Before switching, it snapshots the current repository so the rollback itself is traceable.
func restoreUserFileVersion(ctx context.Context, sess *xorm.Session, userIdentity string, fileIdentity string, versionIdentity string) (*models.UserRepository, *models.RepositoryPool, error) {
	file, err := getUserRepository(ctx, sess, userIdentity, fileIdentity, false, true)
	if err != nil {
		return nil, nil, err
	}
	if file.IsDir == 1 {
		return nil, nil, errors.New(ctx, "folder does not support version restore", nil, map[string]interface{}{
			"file_identity": fileIdentity,
		})
	}

	version := new(models.UserFileVersion)
	has, err := sess.SQL(
		"SELECT * FROM user_file_version WHERE identity = ? AND user_identity = ? AND file_identity = ? LIMIT 1 FOR UPDATE",
		versionIdentity,
		userIdentity,
		file.Identity,
	).Get(version)
	if err != nil {
		return nil, nil, errors.New(ctx, "query file version failed", err, map[string]interface{}{
			"file_identity":    fileIdentity,
			"version_identity": versionIdentity,
		})
	}
	if !has {
		return nil, nil, errors.New(ctx, "file version does not exist", nil, map[string]interface{}{
			"file_identity":    fileIdentity,
			"version_identity": versionIdentity,
		})
	}

	currentRepository := new(models.RepositoryPool)
	has, err = sess.Where("identity = ?", file.RepositoryIdentity).Get(currentRepository)
	if err != nil {
		return nil, nil, errors.New(ctx, "query current repository failed", err, map[string]interface{}{
			"file_identity": fileIdentity,
		})
	}
	if !has {
		return nil, nil, errors.New(ctx, "current repository does not exist", nil, map[string]interface{}{
			"file_identity": fileIdentity,
		})
	}

	targetRepository := new(models.RepositoryPool)
	has, err = sess.Where("identity = ?", version.RepositoryIdentity).Get(targetRepository)
	if err != nil {
		return nil, nil, errors.New(ctx, "query target repository failed", err, map[string]interface{}{
			"file_identity":       fileIdentity,
			"repository_identity": version.RepositoryIdentity,
		})
	}
	if !has {
		return nil, nil, errors.New(ctx, "target repository does not exist", nil, map[string]interface{}{
			"file_identity":       fileIdentity,
			"repository_identity": version.RepositoryIdentity,
		})
	}

	// Restoring the same repository is treated as a cheap idempotent success.
	if currentRepository.Identity == targetRepository.Identity {
		return file, targetRepository, nil
	}

	if err := snapshotUserFileVersion(ctx, sess, file, currentRepository, "restore"); err != nil {
		return nil, nil, err
	}

	restoreExt := strings.TrimSpace(version.Ext)
	if restoreExt == "" {
		restoreExt = strings.TrimSpace(targetRepository.Ext)
	}
	if restoreExt == "" {
		restoreExt = file.Ext
	}

	if _, err := sess.Exec(
		"UPDATE user_repository SET repository_identity = ?, ext = ?, updated_at = NOW() WHERE identity = ? AND user_identity = ? AND deleted_at IS NULL",
		targetRepository.Identity,
		restoreExt,
		file.Identity,
		userIdentity,
	); err != nil {
		return nil, nil, errors.New(ctx, "restore file version failed", err, map[string]interface{}{
			"file_identity":       fileIdentity,
			"repository_identity": targetRepository.Identity,
			"version_identity":    versionIdentity,
		})
	}

	file.RepositoryIdentity = targetRepository.Identity
	file.Ext = restoreExt
	return file, targetRepository, nil
}
