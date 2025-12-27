@echo off
chcp 65001 >nul 2>&1
title superAIAutoCutVideo Setup
color 0A

echo.
echo  ============================================================
echo       superAIAutoCutVideo Installation Script (Windows)
echo       Using China Mirror Sources - No VPN Required
echo  ============================================================
echo.

:: China mirror configuration
set PIP_MIRROR=https://pypi.tuna.tsinghua.edu.cn/simple
set NPM_MIRROR=https://registry.npmmirror.com

:: ========== Environment Check ==========
echo [Environment Check]
echo.

:: Check Python
echo   Checking Python...
where python >nul 2>&1
if errorlevel 1 (
    color 0C
    echo   [X] Python not found
    echo.
    echo   Please install Python 3.10+ first:
    echo   Download: https://www.python.org/downloads/
    echo.
    echo   IMPORTANT: Check "Add Python to PATH" during installation!
    echo.
    pause
    exit /b 1
)
for /f "tokens=2" %%i in ('python --version 2^>^&1') do set PYVER=%%i
echo   [OK] Python %PYVER%

:: Check Node.js
echo   Checking Node.js...
where node >nul 2>&1
if errorlevel 1 (
    color 0C
    echo   [X] Node.js not found
    echo.
    echo   Please install Node.js 18+ first:
    echo   Download: https://nodejs.org/
    echo.
    pause
    exit /b 1
)
for /f %%i in ('node --version 2^>^&1') do set NODEVER=%%i
echo   [OK] Node.js %NODEVER%

echo   [OK] Using China mirrors (No VPN needed)
echo       pip: Tsinghua University
echo       npm: Taobao Mirror

echo.
echo ============================================================
echo.

:: ========== Step 1: Create Virtual Environment ==========
echo [1/6] Creating Python virtual environment...
cd /d "%~dp0..\backend"
if exist "venv\Scripts\activate.bat" (
    echo       Virtual environment exists, skipping
) else (
    echo       Creating venv...
    python -m venv venv
    if errorlevel 1 (
        color 0C
        echo       [X] Failed to create virtual environment
        pause
        exit /b 1
    )
    echo       [OK] Virtual environment created
)
echo.

:: ========== Step 2: Configure pip mirror and upgrade ==========
echo [2/6] Configuring pip China mirror...
call venv\Scripts\activate.bat
python -m pip config set global.index-url %PIP_MIRROR% >nul 2>&1
python -m pip install --upgrade pip -i %PIP_MIRROR% --quiet
echo       [OK] pip configured with Tsinghua mirror
echo.

:: ========== Step 3: Install backend dependencies ==========
echo [3/6] Installing Python dependencies...
echo       This may take a few minutes, please wait...
pip install -r requirements.txt -i %PIP_MIRROR%
if errorlevel 1 (
    color 0E
    echo       [!] Some dependencies failed, check network
) else (
    echo       [OK] Backend dependencies installed
)
echo.

:: ========== Step 4: Install IndexTTS2 ==========
echo [4/6] Installing IndexTTS2...
echo       Trying Gitee mirror (China)...

:: Try Gitee mirror first
pip install git+https://gitee.com/mirrors/index-tts.git -i %PIP_MIRROR% 2>nul
if errorlevel 1 (
    echo       Gitee failed, trying GitHub...
    pip install git+https://github.com/index-tts/index-tts.git -i %PIP_MIRROR% 2>nul
    if errorlevel 1 (
        color 0E
        echo       [!] IndexTTS2 installation failed
        echo.
        echo       Manual install method:
        echo       1. Download: https://github.com/index-tts/index-tts/archive/refs/heads/main.zip
        echo       2. Extract and run: pip install .
        echo.
    ) else (
        echo       [OK] IndexTTS2 installed (GitHub)
    )
) else (
    echo       [OK] IndexTTS2 installed (Gitee mirror)
)
echo.

:: ========== Step 5: Install frontend dependencies ==========
echo [5/6] Installing npm dependencies...
cd /d "%~dp0..\frontend"
echo       Configuring npm Taobao mirror...
call npm config set registry %NPM_MIRROR%
echo       This may take a few minutes, please wait...
call npm install
if errorlevel 1 (
    color 0E
    echo       [!] Frontend dependencies failed
) else (
    echo       [OK] Frontend dependencies installed
)
echo.

:: ========== Step 6: Download model ==========
echo [6/6] Downloading IndexTTS2 model...
cd /d "%~dp0..\backend"
call venv\Scripts\activate.bat

:: Check if model exists
if exist "checkpoints\config.yaml" (
    echo       Model already exists, skipping
    goto :install_done
)

echo       Model size: 3-5GB, please ensure network is stable...
echo       Downloading from ModelScope (China server, no VPN)...
echo.

:: Install modelscope first
pip install modelscope -i %PIP_MIRROR% --quiet

python -c "from modelscope import snapshot_download; snapshot_download('IndexTeam/IndexTTS-1.5', local_dir='./checkpoints')"
if errorlevel 1 (
    color 0E
    echo.
    echo       [!] Model download failed
    echo.
    echo       Manual download method:
    echo       1. Visit: https://modelscope.cn/models/IndexTeam/IndexTTS-1.5
    echo       2. Click "Download Model"
    echo       3. Extract to backend\checkpoints\ folder
    echo.
) else (
    echo       [OK] Model downloaded
)

:install_done
echo.
color 0A
echo  ============================================================
echo                    Installation Complete!
echo  ============================================================
echo.
echo  How to start:
echo    Double-click scripts\start.bat
echo.
echo  Or manually:
echo    Backend: cd backend ^&^& venv\Scripts\activate ^&^& python main.py
echo    Frontend: cd frontend ^&^& npm run dev
echo.
echo  Note:
echo    1. Prepare 6 voice WAV files for first use
echo       Location: backend\serviceData\index_tts\voices\
echo    2. See docs\IndexTTS2_Guide.md for details
echo.
echo  Mirrors used:
echo    pip:   Tsinghua University (%PIP_MIRROR%)
echo    npm:   Taobao Mirror (%NPM_MIRROR%)
echo    Model: ModelScope (Aliyun China)
echo.
echo ============================================================
pause
