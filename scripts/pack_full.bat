@echo off

:: Prevent flash close
if "%~1"=="" (
    start cmd /k "%~f0" run
    exit /b
)

setlocal enabledelayedexpansion
chcp 65001 >nul 2>&1
title Create Full Package
color 0D

echo.
echo  ============================================================
echo       Create Full Distribution Package
echo       (Includes venv + models, 5-10GB)
echo  ============================================================
echo.

cd /d "%~dp0.."

:: Check components
echo   Checking components...
echo.

set READY=1

if not exist "backend\venv\Scripts\activate.bat" (
    echo   [X] Virtual environment not found
    echo       Run install.bat first
    set READY=0
)

if not exist "backend\checkpoints\gpt.pth" (
    echo   [X] Model files not found
    echo       Run download_model.bat first
    set READY=0
)

if not exist "frontend\node_modules" (
    echo   [X] Frontend packages not found
    echo       Run install.bat first
    set READY=0
)

if "!READY!"=="0" (
    echo.
    echo   Please complete installation first!
    goto :end
)

echo   [OK] All components ready
echo.

:: Get timestamp using PowerShell (more compatible)
for /f "tokens=*" %%a in ('powershell -command "Get-Date -Format 'yyyyMMdd_HHmm'"') do set "TIMESTAMP=%%a"
if "%TIMESTAMP%"=="" set "TIMESTAMP=package"

set "PACKAGE_NAME=superAIAutoCutVideo_full_%TIMESTAMP%"
set "PACKAGE_DIR=%USERPROFILE%\Desktop\%PACKAGE_NAME%"

echo   Package name: %PACKAGE_NAME%
echo   Location: %USERPROFILE%\Desktop\
echo.
echo   This will take 5-15 minutes...
echo.
pause

:: Create package directory
echo.
echo   [1/6] Creating package directory...
if exist "%PACKAGE_DIR%" rd /s /q "%PACKAGE_DIR%"
mkdir "%PACKAGE_DIR%"
if not exist "%PACKAGE_DIR%" (
    echo   [X] Failed to create directory
    goto :end
)

:: Copy backend
echo   [2/6] Copying backend (this takes a while)...
xcopy /E /I /Q /Y "backend" "%PACKAGE_DIR%\backend"
echo       Done

:: Copy entire frontend (except node_modules)
echo   [3/6] Copying frontend...
xcopy /E /I /Q /Y "frontend" "%PACKAGE_DIR%\frontend" /EXCLUDE:scripts\pack_exclude_frontend.txt
echo       Done

:: Copy scripts
echo   [4/6] Copying scripts...
xcopy /E /I /Q /Y "scripts" "%PACKAGE_DIR%\scripts"
echo       Done

:: Copy docs
echo   [5/6] Copying documentation...
copy /Y "*.md" "%PACKAGE_DIR%\" >nul 2>&1
copy /Y "LICENSE" "%PACKAGE_DIR%\" >nul 2>&1
echo       Done

:: Create exclude file if not exists
echo   [6/6] Finalizing...
echo       Done

echo.
echo  ============================================================
echo                    Package Created!
echo  ============================================================
echo.
echo   Location: %PACKAGE_DIR%
echo.
echo   Size: Check folder properties
echo.
echo   To create ZIP:
echo   1. Right-click folder on Desktop
echo   2. Send to - Compressed (zipped) folder
echo   Or use 7-Zip for smaller size
echo.

:end
echo.
echo Press any key to exit...
pause >nul
endlocal
