@echo off
setlocal enabledelayedexpansion
chcp 65001 >nul 2>&1
title superAIAutoCutVideo Setup
color 0A

echo.
echo  ============================================================
echo       superAIAutoCutVideo Installation Script
echo       Auto-install with winget / China mirror links
echo  ============================================================
echo.

:: China mirror configuration
set PIP_MIRROR=https://pypi.tuna.tsinghua.edu.cn/simple
set NPM_MIRROR=https://registry.npmmirror.com

:: ========== Check winget ==========
set WINGET_OK=0
where winget >nul 2>&1
if not errorlevel 1 (
    set WINGET_OK=1
    echo   [OK] winget available
)
echo.

:: ========== Check Python ==========
echo [Checking Python...]
set PYTHON_OK=0
where python >nul 2>&1
if not errorlevel 1 (
    :: Verify it's real Python, not Windows Store alias
    python --version >nul 2>&1
    if not errorlevel 1 (
        for /f "tokens=*" %%i in ('python --version 2^>^&1') do set PYVER=%%i
        echo   [OK] !PYVER!
        set PYTHON_OK=1
    )
)

if "!PYTHON_OK!"=="0" (
    echo   [X] Python not found or not working
    echo.
    
    :: Try winget install
    if "!WINGET_OK!"=="1" (
        echo   Trying winget install...
        winget install Python.Python.3.11 --accept-package-agreements --accept-source-agreements >nul 2>&1
        if not errorlevel 1 (
            echo   [OK] Python installed via winget
            echo.
            echo   *** Please CLOSE this window and run install.bat AGAIN ***
            echo.
            pause
            exit /b 0
        )
    )
    
    :: Manual install instructions
    color 0C
    echo   ============================================
    echo   Please install Python manually:
    echo.
    echo   China Mirror (Fast):
    echo   https://mirrors.huaweicloud.com/python/3.11.9/python-3.11.9-amd64.exe
    echo.
    echo   Official:
    echo   https://www.python.org/downloads/
    echo.
    echo   IMPORTANT: Check [x] Add Python to PATH
    echo   ============================================
    echo.
    pause
    exit /b 1
)

:: ========== Check Node.js ==========
echo.
echo [Checking Node.js...]
set NODE_OK=0
where node >nul 2>&1
if not errorlevel 1 (
    for /f "tokens=*" %%i in ('node --version 2^>^&1') do set NODEVER=%%i
    echo   [OK] Node.js !NODEVER!
    set NODE_OK=1
)

if "!NODE_OK!"=="0" (
    echo   [X] Node.js not found
    echo.
    
    :: Try winget install
    if "!WINGET_OK!"=="1" (
        echo   Trying winget install...
        winget install OpenJS.NodeJS.LTS --accept-package-agreements --accept-source-agreements >nul 2>&1
        if not errorlevel 1 (
            echo   [OK] Node.js installed via winget
            echo.
            echo   *** Please CLOSE this window and run install.bat AGAIN ***
            echo.
            pause
            exit /b 0
        )
    )
    
    :: Manual install instructions
    color 0C
    echo   ============================================
    echo   Please install Node.js manually:
    echo.
    echo   China Mirror (Fast):
    echo   https://mirrors.huaweicloud.com/nodejs/v20.18.0/node-v20.18.0-x64.msi
    echo.
    echo   Official:
    echo   https://nodejs.org/
    echo   ============================================
    echo.
    pause
    exit /b 1
)

:: ========== Check Git ==========
echo.
echo [Checking Git...]
set GIT_OK=0
where git >nul 2>&1
if not errorlevel 1 (
    echo   [OK] Git available
    set GIT_OK=1
) else (
    echo   [!] Git not found (optional, for IndexTTS2)
    echo.
    echo   China Mirror (Fast):
    echo   https://mirrors.huaweicloud.com/git-for-windows/v2.47.0.windows.2/Git-2.47.0.2-64-bit.exe
    echo.
    echo   Git is optional. Continue without it?
    choice /C YN /M "Continue"
    if errorlevel 2 (
        pause
        exit /b 1
    )
)

echo.
echo   Using China mirrors (No VPN needed)
echo       pip: Tsinghua University
echo       npm: Taobao Mirror
echo       model: ModelScope

echo.
echo ============================================================
echo.

:: ========== Step 1: Create Virtual Environment ==========
echo [1/6] Creating Python virtual environment...
cd /d "%~dp0..\backend"
if exist "venv\Scripts\activate.bat" (
    echo       Already exists, skipping
) else (
    echo       Creating venv...
    python -m venv venv
    if errorlevel 1 (
        color 0C
        echo       [X] Failed to create virtual environment
        echo.
        echo       This usually means Python is not installed correctly.
        echo       Please reinstall Python with "Add to PATH" checked.
        pause
        exit /b 1
    )
    echo       [OK] Created
)
echo.

:: ========== Step 2: Configure pip ==========
echo [2/6] Configuring pip with Tsinghua mirror...
call venv\Scripts\activate.bat
python -m pip config set global.index-url %PIP_MIRROR% >nul 2>&1
python -m pip install --upgrade pip -i %PIP_MIRROR% --quiet
echo       [OK] Done
echo.

:: ========== Step 3: Install backend dependencies ==========
echo [3/6] Installing Python dependencies...
echo       Please wait (2-5 minutes)...
pip install -r requirements.txt -i %PIP_MIRROR%
if errorlevel 1 (
    color 0E
    echo       [!] Some failed, but continuing...
) else (
    echo       [OK] Done
)
echo.

:: ========== Step 4: Install IndexTTS2 ==========
echo [4/6] Installing IndexTTS2...

if "!GIT_OK!"=="0" (
    echo       [SKIP] Git not available
    echo       You can install IndexTTS2 manually later
    goto :skip_indextts
)

echo       Trying Gitee mirror...
pip install git+https://gitee.com/mirrors/index-tts.git -i %PIP_MIRROR% 2>nul
if errorlevel 1 (
    echo       Gitee failed, trying GitHub...
    pip install git+https://github.com/index-tts/index-tts.git -i %PIP_MIRROR% 2>nul
    if errorlevel 1 (
        echo       [!] Failed - you can install manually later
    ) else (
        echo       [OK] Done (GitHub)
    )
) else (
    echo       [OK] Done (Gitee)
)
:skip_indextts
echo.

:: ========== Step 5: Install frontend ==========
echo [5/6] Installing frontend dependencies...
cd /d "%~dp0..\frontend"
call npm config set registry %NPM_MIRROR%
echo       Please wait (1-3 minutes)...
call npm install
if errorlevel 1 (
    color 0E
    echo       [!] Failed
) else (
    echo       [OK] Done
)
echo.

:: ========== Step 6: Download model ==========
echo [6/6] Downloading IndexTTS2 model...
cd /d "%~dp0..\backend"
call venv\Scripts\activate.bat

if exist "checkpoints\config.yaml" (
    echo       Already exists, skipping
    goto :install_done
)

echo       Model: 3-5GB from ModelScope (China)
echo       This may take 10-30 minutes...
echo.

pip install modelscope -i %PIP_MIRROR% --quiet
python -c "from modelscope import snapshot_download; snapshot_download('IndexTeam/IndexTTS-1.5', local_dir='./checkpoints')"
if errorlevel 1 (
    color 0E
    echo.
    echo       [!] Download failed
    echo.
    echo       Manual download:
    echo       https://modelscope.cn/models/IndexTeam/IndexTTS-1.5
    echo       Download all files to: backend\checkpoints\
) else (
    echo       [OK] Done
)

:install_done
echo.
color 0A
echo  ============================================================
echo                    Installation Complete!
echo  ============================================================
echo.
echo  To start: Double-click scripts\start.bat
echo.
echo  Voice files needed for IndexTTS2:
echo    Put 6 WAV files in: backend\serviceData\index_tts\voices\
echo.
echo ============================================================
pause
endlocal
