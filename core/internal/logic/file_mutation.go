package logic

import (
	"cloud_disk/core/internal/errors"
	"context"
	"fmt"

	"xorm.io/xorm"
)

// softDeleteFiles / restoreFiles / hardDeleteFiles
// 是批量删除与回收站功能共用的基础操作。
// 这里统一按 identity 批量处理，避免每个业务逻辑里重复拼 SQL。

func softDeleteFiles(ctx context.Context, sess *xorm.Session, userIdentity string, identities []string) error {
	if len(identities) == 0 {
		return nil
	}
	args := make([]interface{}, 0, len(identities)+1)
	args = append(args, userIdentity)
	for _, identity := range identities {
		args = append(args, identity)
	}
	sql := fmt.Sprintf(
		"UPDATE user_repository SET deleted_at = NOW() WHERE user_identity = ? AND identity IN (%s) AND deleted_at IS NULL",
		placeholders(len(identities)),
	)
	execArgs := append([]interface{}{sql}, args...)
	if _, err := sess.Exec(execArgs...); err != nil {
		return errors.New(ctx, "move files to recycle bin failed", err, map[string]interface{}{
			"count": len(identities),
		})
	}
	return nil
}

func restoreFiles(ctx context.Context, sess *xorm.Session, userIdentity string, identities []string) error {
	if len(identities) == 0 {
		return nil
	}
	args := make([]interface{}, 0, len(identities)+1)
	args = append(args, userIdentity)
	for _, identity := range identities {
		args = append(args, identity)
	}
	sql := fmt.Sprintf(
		"UPDATE user_repository SET deleted_at = NULL WHERE user_identity = ? AND identity IN (%s)",
		placeholders(len(identities)),
	)
	execArgs := append([]interface{}{sql}, args...)
	if _, err := sess.Exec(execArgs...); err != nil {
		return errors.New(ctx, "restore files failed", err, map[string]interface{}{
			"count": len(identities),
		})
	}
	return nil
}

func hardDeleteFiles(ctx context.Context, sess *xorm.Session, userIdentity string, identities []string) error {
	if len(identities) == 0 {
		return nil
	}
	args := make([]interface{}, 0, len(identities)+1)
	args = append(args, userIdentity)
	for _, identity := range identities {
		args = append(args, identity)
	}
	sql := fmt.Sprintf(
		"DELETE FROM user_repository WHERE user_identity = ? AND identity IN (%s)",
		placeholders(len(identities)),
	)
	execArgs := append([]interface{}{sql}, args...)
	if _, err := sess.Exec(execArgs...); err != nil {
		return errors.New(ctx, "delete files permanently failed", err, map[string]interface{}{
			"count": len(identities),
		})
	}
	return nil
}
