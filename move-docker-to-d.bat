@echo off
echo ========================================
echo Move Docker Desktop to D Drive
echo ========================================
echo.
echo This script will:
echo 1. Stop Docker Desktop
echo 2. Export WSL2 Docker data
echo 3. Move to D:\Docker
echo 4. Re-import and set as default
echo.
echo Press Ctrl+C to cancel, or
pause

echo.
echo Step 1: Stopping Docker Desktop...
wsl --shutdown
timeout /t 3

echo.
echo Step 2: Creating D:\Docker directory...
if not exist "D:\Docker" mkdir "D:\Docker"

echo.
echo Step 3: Exporting docker-desktop-data...
wsl --export docker-desktop-data "D:\Docker\docker-desktop-data.tar"

echo.
echo Step 4: Exporting docker-desktop...
wsl --export docker-desktop "D:\Docker\docker-desktop.tar"

echo.
echo Step 5: Unregistering old distributions...
wsl --unregister docker-desktop-data
wsl --unregister docker-desktop

echo.
echo Step 6: Importing to D:\Docker...
wsl --import docker-desktop-data "D:\Docker\data" "D:\Docker\docker-desktop-data.tar" --version 2
wsl --import docker-desktop "D:\Docker\desktop" "D:\Docker\docker-desktop.tar" --version 2

echo.
echo ========================================
echo Migration completed!
echo ========================================
echo.
echo Docker data is now at: D:\Docker
echo.
echo You can now:
echo 1. Start Docker Desktop
echo 2. Delete the .tar files to free up space:
echo    - D:\Docker\docker-desktop-data.tar
echo    - D:\Docker\docker-desktop.tar
echo.
pause
