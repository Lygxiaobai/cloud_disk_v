package logic

import (
	"cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/models"
	"context"
	"strings"
	"time"

	"xorm.io/xorm"
)

func getShareBasic(ctx context.Context, sess *xorm.Session, shareIdentity string, includeDeleted bool, forUpdate bool) (*models.ShareBasic, error) {
	row := new(models.ShareBasic)
	sql := "SELECT * FROM share_basic WHERE identity = ?"
	if !includeDeleted {
		sql += " AND deleted_at IS NULL"
	}
	sql += " LIMIT 1"
	if forUpdate {
		sql += " FOR UPDATE"
	}

	has, err := sess.SQL(sql, shareIdentity).Get(row)
	if err != nil {
		return nil, errors.Internal(ctx, "查询分享记录失败", err, map[string]interface{}{
			"share_identity": shareIdentity,
		})
	}
	if !has {
		return nil, errors.NotFound(ctx, errors.CodeShareNotFound, "分享不存在", nil, map[string]interface{}{
			"share_identity": shareIdentity,
		})
	}
	return row, nil
}

func validateShareAccess(ctx context.Context, sess *xorm.Session, shareIdentity string, accessCode string, forUpdate bool) (*models.ShareBasic, error) {
	share, err := getShareBasic(ctx, sess, shareIdentity, false, forUpdate)
	if err != nil {
		return nil, err
	}

	if share.ExpiredTime > 0 && share.CreatedAt.Add(time.Duration(share.ExpiredTime)*time.Second).Before(time.Now()) {
		return nil, errors.InvalidParam(ctx, "分享已过期", nil, map[string]interface{}{
			"share_identity": shareIdentity,
		})
	}

	expectedCode := strings.TrimSpace(share.AccessCode)
	if expectedCode != "" && !strings.EqualFold(expectedCode, strings.TrimSpace(accessCode)) {
		return nil, errors.VerificationFailed(ctx, "分享提取码错误", nil, map[string]interface{}{
			"share_identity": shareIdentity,
		})
	}

	return share, nil
}
