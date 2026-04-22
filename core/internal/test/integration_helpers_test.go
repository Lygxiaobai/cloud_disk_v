package test

import (
	"context"
	"os"
	"testing"
	"time"

	"cloud_disk/core/internal/config"
	"cloud_disk/core/internal/models"
	"cloud_disk/core/internal/svc"

	"github.com/stretchr/testify/require"
)

const runIntegrationEnv = "RUN_INTEGRATION_TESTS"

type integrationConfig struct {
	mysqlDSN      string
	redisAddr     string
	redisPassword string
	redisDB       int
}

func requireIntegration(t *testing.T) integrationConfig {
	t.Helper()

	if os.Getenv(runIntegrationEnv) != "1" {
		t.Skip("set RUN_INTEGRATION_TESTS=1 to run integration tests")
	}

	cfg := integrationConfig{
		mysqlDSN:      envOr("MYSQL_DSN_TEST", "root:123456@tcp(127.0.0.1:3307)/cloud_disk?charset=utf8mb4&parseTime=True&loc=Local"),
		redisAddr:     envOr("REDIS_ADDR_TEST", "127.0.0.1:6380"),
		redisPassword: os.Getenv("REDIS_PASSWORD_TEST"),
		redisDB:       9,
	}

	engine := models.Init(cfg.mysqlDSN)
	require.NotNil(t, engine, "mysql init failed")
	require.NoError(t, engine.Ping())
	require.NoError(t, engine.Close())

	rdb := models.InitRedis(cfg.redisAddr, cfg.redisPassword, cfg.redisDB)
	require.NoError(t, rdb.Ping(context.Background()).Err())
	require.NoError(t, rdb.Close())

	return cfg
}

func newIntegrationServiceContext(t *testing.T, cfg integrationConfig) *svc.ServiceContext {
	t.Helper()

	c := config.Config{}
	c.Mysql.DataSource = cfg.mysqlDSN
	c.Redis.Addr = cfg.redisAddr
	c.Redis.Password = cfg.redisPassword
	c.Redis.DB = cfg.redisDB
	c.JWT.AccessSecret = "integration-access-secret"
	c.JWT.RefreshSecret = "integration-refresh-secret"
	c.JWT.AccessExpire = 3600
	c.JWT.RefreshExpire = 7200
	c.RateLimit.LoginLockThreshold = 5
	c.RateLimit.LoginLockMinutes = 15
	c.Mail.CodeExpire = 300
	c.Mail.CodeLen = 6

	engine := models.Init(cfg.mysqlDSN)
	require.NotNil(t, engine)
	rdb := models.InitRedis(cfg.redisAddr, cfg.redisPassword, cfg.redisDB)
	require.NoError(t, rdb.FlushDB(context.Background()).Err())

	t.Cleanup(func() {
		_ = rdb.FlushDB(context.Background()).Err()
		_ = rdb.Close()
		_ = engine.Close()
	})

	return &svc.ServiceContext{
		Config: c,
		Engine: engine,
		RDB:    rdb,
	}
}

func envOr(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func uniqueSuffix() string {
	return time.Now().Format("20060102150405.000000000")
}
