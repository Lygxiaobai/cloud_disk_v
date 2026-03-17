@echo off
echo ========================================
echo Starting Elasticsearch with Docker
echo ========================================
echo.

echo Checking Docker installation...
docker --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Docker is not installed or not running
    echo Please install Docker Desktop: https://www.docker.com/products/docker-desktop
    pause
    exit /b 1
)

echo.
echo Starting Elasticsearch container...
docker run -d --name elasticsearch -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" -e "xpack.security.enabled=false" elasticsearch:8.11.0

if %errorlevel% equ 0 (
    echo.
    echo ========================================
    echo Elasticsearch started successfully!
    echo ========================================
    echo.
    echo Access URL: http://localhost:9200
    echo.
    echo View logs: docker logs -f elasticsearch
    echo Stop service: docker stop elasticsearch
    echo Remove container: docker rm elasticsearch
    echo.
) else (
    echo.
    echo Failed to start. Container may already exist.
    echo Trying to start existing container...
    docker start elasticsearch

    if %errorlevel% equ 0 (
        echo Elasticsearch container started
    ) else (
        echo Please check manually: docker ps -a
    )
)

pause
