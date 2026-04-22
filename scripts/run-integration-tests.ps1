$composeFile = "deploy\docker-compose.integration.yml"
$mysqlPort = if ($env:MYSQL_PORT) { $env:MYSQL_PORT } else { "3307" }
$redisPort = if ($env:REDIS_PORT) { $env:REDIS_PORT } else { "6380" }

try {
  docker compose -f $composeFile up -d
  if ($LASTEXITCODE -ne 0) {
    throw "failed to start integration dependencies"
  }

  for ($i = 0; $i -lt 30; $i++) {
    docker exec cloud-disk-mysql-test mysqladmin ping -h 127.0.0.1 -p123456 --silent *> $null
    $mysqlReady = $LASTEXITCODE -eq 0

    $redisOutput = docker exec cloud-disk-redis-test redis-cli ping 2>$null
    $redisReady = $LASTEXITCODE -eq 0 -and $redisOutput -match "PONG"

    if ($mysqlReady -and $redisReady) {
      break
    }

    Start-Sleep -Seconds 5
    if ($i -eq 29) {
      throw "integration dependencies did not become healthy"
    }
  }

  $env:RUN_INTEGRATION_TESTS = "1"
  $env:MYSQL_DSN_TEST = "root:123456@tcp(127.0.0.1:$mysqlPort)/cloud_disk?charset=utf8mb4&parseTime=True&loc=Local"
  $env:REDIS_ADDR_TEST = "127.0.0.1:$redisPort"

  go test ./core/internal/test -count=1
}
finally {
  docker compose -f $composeFile down -v *> $null
}
