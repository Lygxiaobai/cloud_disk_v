@echo off
echo ========================================
echo Starting RabbitMQ with Docker
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
echo Starting RabbitMQ container...
docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management

if %errorlevel% equ 0 (
    echo.
    echo ========================================
    echo RabbitMQ started successfully!
    echo ========================================
    echo.
    echo Management UI: http://localhost:15672
    echo Username: guest
    echo Password: guest
    echo.
    echo View logs: docker logs -f rabbitmq
    echo Stop service: docker stop rabbitmq
    echo Remove container: docker rm rabbitmq
    echo.
) else (
    echo.
    echo Failed to start. Container may already exist.
    echo Trying to start existing container...
    docker start rabbitmq

    if %errorlevel% equ 0 (
        echo RabbitMQ container started
    ) else (
        echo Please check manually: docker ps -a
    )
)

pause
