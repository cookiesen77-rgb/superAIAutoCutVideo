@echo off
chcp 65001 >nul 2>&1
title Pack Full Version (with venv and model)
color 0E

echo.
echo  ============================================================
echo       Pack Full Version - Including venv and model
echo       Creates ready-to-use ZIP package
echo  ============================================================
echo.

set PROJECT_DIR=%~dp0..
set PACK_NAME=superAIAutoCutVideo_Full
set PACK_DIR=%USERPROFILE%\Desktop\%PACK_NAME%

:: ========== Step 1: Check dependencies ==========
echo [1/5] Checking dependencies...
cd /d "%PROJECT_DIR%"

:: If venv doesn't exist, run install first
if not exist "backend\venv\Scripts\activate.bat" (
    echo       Virtual environment not found, running install...
    call scripts\install.bat
    if errorlevel 1 (
        echo       [X] Installation failed
        pause
        exit /b 1
    )
)
echo       [OK] Dependencies installed
echo.

:: ========== Step 2: Check model ==========
echo [2/5] Checking model files...
if not exist "backend\checkpoints\config.yaml" (
    echo       [!] Model files not found
    echo       Please run install.bat to download model first
    echo       Or manually download to backend\checkpoints\
    pause
    exit /b 1
)
echo       [OK] Model files ready
echo.

:: ========== Step 3: Check voice files ==========
echo [3/5] Checking voice files...
set VOICE_DIR=backend\serviceData\index_tts\voices
set VOICE_COUNT=0
if exist "%VOICE_DIR%\male_youth.wav" set /a VOICE_COUNT+=1
if exist "%VOICE_DIR%\male_mature.wav" set /a VOICE_COUNT+=1
if exist "%VOICE_DIR%\female_sweet.wav" set /a VOICE_COUNT+=1
if exist "%VOICE_DIR%\female_elegant.wav" set /a VOICE_COUNT+=1
if exist "%VOICE_DIR%\female_lively.wav" set /a VOICE_COUNT+=1
if exist "%VOICE_DIR%\child.wav" set /a VOICE_COUNT+=1

if %VOICE_COUNT% LSS 6 (
    echo       [!] Voice files incomplete (%VOICE_COUNT%/6)
    echo       Put 6 voice WAV files in: %VOICE_DIR%\
    echo.
    echo       Missing files will cause voice unavailable. Continue?
    choice /C YN /M "Continue packing"
    if errorlevel 2 exit /b 1
) else (
    echo       [OK] Voice files complete (6/6)
)
echo.

:: ========== Step 4: Create pack directory ==========
echo [4/5] Creating pack directory...
if exist "%PACK_DIR%" (
    echo       Cleaning old pack directory...
    rmdir /s /q "%PACK_DIR%"
)
mkdir "%PACK_DIR%"

echo       Copying project files...
:: Copy backend (with venv and model)
xcopy /E /I /Q "backend" "%PACK_DIR%\backend" >nul
:: Remove __pycache__
for /d /r "%PACK_DIR%\backend" %%d in (__pycache__) do @if exist "%%d" rmdir /s /q "%%d"

:: Copy frontend (without node_modules)
mkdir "%PACK_DIR%\frontend"
xcopy /E /I /Q "frontend\src" "%PACK_DIR%\frontend\src" >nul
xcopy /E /I /Q "frontend\public" "%PACK_DIR%\frontend\public" >nul 2>nul
copy "frontend\package.json" "%PACK_DIR%\frontend\" >nul
copy "frontend\package-lock.json" "%PACK_DIR%\frontend\" >nul 2>nul
copy "frontend\vite.config.ts" "%PACK_DIR%\frontend\" >nul 2>nul
copy "frontend\tsconfig.json" "%PACK_DIR%\frontend\" >nul 2>nul
copy "frontend\tsconfig.node.json" "%PACK_DIR%\frontend\" >nul 2>nul
copy "frontend\index.html" "%PACK_DIR%\frontend\" >nul 2>nul
copy "frontend\tailwind.config.js" "%PACK_DIR%\frontend\" >nul 2>nul
copy "frontend\postcss.config.js" "%PACK_DIR%\frontend\" >nul 2>nul

:: Copy scripts and docs
xcopy /E /I /Q "scripts" "%PACK_DIR%\scripts" >nul
xcopy /E /I /Q "docs" "%PACK_DIR%\docs" >nul 2>nul
copy "LICENSE" "%PACK_DIR%\" >nul 2>nul
copy "README.md" "%PACK_DIR%\" >nul 2>nul

echo       [OK] Files copied
echo.

:: ========== Step 5: Create startup scripts ==========
echo [5/5] Creating startup scripts...

:: Create simplified startup script (run directly without install)
(
echo @echo off
echo chcp 65001 ^>nul 2^>^&1
echo title superAIAutoCutVideo
echo color 0A
echo.
echo echo  Starting superAIAutoCutVideo...
echo echo.
echo.
echo :: Start backend
echo cd /d "%%~dp0backend"
echo start "Backend" cmd /k "call venv\Scripts\activate.bat && python main.py"
echo.
echo :: Wait for backend
echo timeout /t 5 /nobreak ^>nul
echo.
echo :: Start frontend
echo cd /d "%%~dp0frontend"
echo start "Frontend" cmd /k "npm run dev"
echo.
echo echo.
echo echo  ============================================
echo echo    Services Started!
echo echo    Backend:  http://localhost:8000
echo echo    Frontend: http://localhost:5173
echo echo  ============================================
echo echo.
echo echo  To stop: Close both command windows
echo pause
) > "%PACK_DIR%\START.bat"

:: Create frontend dependency install script (run once for first use)
(
echo @echo off
echo chcp 65001 ^>nul 2^>^&1
echo title Install Frontend Dependencies
echo echo Installing frontend dependencies...
echo echo This only needs to run once.
echo echo.
echo cd /d "%%~dp0frontend"
echo call npm config set registry https://registry.npmmirror.com
echo call npm install
echo echo.
echo echo [OK] Frontend dependencies installed!
echo echo Now you can double-click START.bat to run the app.
echo pause
) > "%PACK_DIR%\FIRST_TIME_SETUP.bat"

:: Create README
(
echo ============================================
echo   superAIAutoCutVideo - Quick Start Guide
echo ============================================
echo.
echo FOR FIRST TIME USE:
echo   1. Double-click FIRST_TIME_SETUP.bat
echo      ^(This installs frontend dependencies, ~1 min^)
echo.
echo TO START THE APP:
echo   2. Double-click START.bat
echo.
echo ============================================
) > "%PACK_DIR%\README.txt"

echo       [OK] Startup scripts created
echo.

:: ========== Show results ==========
echo ============================================================
echo.
echo  Packing Complete!
echo.
echo  Output: %PACK_DIR%
echo.
echo  Structure:
echo    %PACK_NAME%\
echo    +-- backend\          (with venv and checkpoints)
echo    +-- frontend\         (without node_modules)
echo    +-- scripts\
echo    +-- docs\
echo    +-- START.bat         ^<-- Double-click to start
echo    +-- FIRST_TIME_SETUP.bat
echo    +-- README.txt
echo.
echo ============================================================
echo.
echo  Customer usage:
echo    1. First time: Double-click FIRST_TIME_SETUP.bat
echo    2. Start app:  Double-click START.bat
echo.
echo  Compress to ZIP?
choice /C YN /M "Create ZIP"
if errorlevel 2 goto :done

echo.
echo  Compressing...
cd /d "%USERPROFILE%\Desktop"
powershell -Command "Compress-Archive -Path '%PACK_NAME%' -DestinationPath '%PACK_NAME%.zip' -Force"
echo.
echo  [OK] ZIP created: %USERPROFILE%\Desktop\%PACK_NAME%.zip
echo.

:done
echo ============================================================
pause
