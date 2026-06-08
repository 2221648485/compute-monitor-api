@echo off
chcp 65001 > nul
setlocal

set "ROOT_DIR=%~dp0"
set "ROOT_DIR=%ROOT_DIR:~0,-1%"
set "IMAGE_NAME=compute-monitor-api"
set "IMAGE_VERSION=1.0.0"
set "IMAGE_TAG=%IMAGE_NAME%:%IMAGE_VERSION%"
set "DEPLOY_DIR=%ROOT_DIR%\deploy"
set "IMAGE_DIR=%DEPLOY_DIR%\images"
set "IMAGE_TAR=%IMAGE_DIR%\%IMAGE_NAME%-%IMAGE_VERSION%.tar"
set "COMPOSE_FILE=%ROOT_DIR%\deployments\docker-compose\docker-compose.yml"

if not exist "%ROOT_DIR%\go.mod" goto bad_root
if not exist "%ROOT_DIR%\Dockerfile" goto bad_root
if not exist "%COMPOSE_FILE%" goto bad_compose

echo ======================== Start build compute-monitor-api ========================
echo Root: %ROOT_DIR%
echo Image: %IMAGE_TAG%
echo.

pushd "%ROOT_DIR%"

echo [STEP] docker build %IMAGE_TAG%
call docker build -t "%IMAGE_TAG%" .
if errorlevel 1 goto fail_pop

echo.
echo [STEP] validate docker compose
call docker compose -f "%COMPOSE_FILE%" config --quiet
if errorlevel 1 goto fail_pop

echo.
echo [STEP] docker save image to tar
if not exist "%IMAGE_DIR%" mkdir "%IMAGE_DIR%"
call docker save -o "%IMAGE_TAR%" "%IMAGE_TAG%"
if errorlevel 1 goto fail_pop

echo.
echo ======================== Build finished ========================
echo Image: %IMAGE_TAG%
echo Tar: %IMAGE_TAR%
echo.
popd
pause
exit /b 0

:fail_pop
popd
goto fail

:bad_root
echo [ERROR] This is not compute-monitor-api project root.
echo [ERROR] go.mod and Dockerfile are required.
pause
exit /b 1

:bad_compose
echo [ERROR] docker compose file not found:
echo %COMPOSE_FILE%
pause
exit /b 1

:fail
echo.
echo [ERROR] Build failed.
pause
exit /b 1
