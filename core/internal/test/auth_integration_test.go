package test

import (
	"context"
	"testing"

	appHelper "cloud_disk/core/internal/helper"
	"cloud_disk/core/internal/logic"
	"cloud_disk/core/internal/models"
	"cloud_disk/core/internal/types"

	"github.com/stretchr/testify/require"
)

func TestRegisterLoginRefreshIntegration(t *testing.T) {
	cfg := requireIntegration(t)
	svcCtx := newIntegrationServiceContext(t, cfg)
	ctx := context.Background()

	suffix := uniqueSuffix()
	userName := "user_" + suffix
	email := "user_" + suffix + "@example.com"
	code := "123456"

	require.NoError(t, svcCtx.RDB.Set(ctx, "code:"+email, code, 0).Err())

	registerLogic := logic.NewUserRegisterLogic(ctx, svcCtx)
	_, err := registerLogic.UserRegister(&types.UserRegisterRequest{
		Name:     userName,
		Password: "secret-123",
		Email:    email,
		Code:     code,
	})
	require.NoError(t, err)

	var user models.UserBasic
	has, err := svcCtx.Engine.Where("name = ? AND email = ?", userName, email).Get(&user)
	require.NoError(t, err)
	require.True(t, has)
	require.NotEqual(t, "secret-123", user.Password)
	require.NoError(t, appHelper.CheckPassword(user.Password, "secret-123"))

	loginLogic := logic.NewUserLoginLogic(ctx, svcCtx)
	loginResp, err := loginLogic.UserLogin(&types.LoginRequest{
		Name:     userName,
		Password: "secret-123",
	})
	require.NoError(t, err)
	require.NotEmpty(t, loginResp.Token)
	require.NotEmpty(t, loginResp.RefreshToken)

	refreshLogic := logic.NewRefreshTokenLogic(ctx, svcCtx)
	refreshResp, err := refreshLogic.RefreshToken(&types.RefreshTokenRequest{
		RefreshToken: loginResp.RefreshToken,
	})
	require.NoError(t, err)
	require.NotEmpty(t, refreshResp.Token)
	require.NotEmpty(t, refreshResp.RefreshToken)
}
