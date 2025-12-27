@echo off

:: Prevent flash close - catch all errors
if "%~1"=="" (
    cmd /k "%~f0" run
    exit /b
)

setlocal enabledelayedexpansion
chcp 65001 >nul 2>&1
title superAIAutoCutVideo Setup
color 0A

echo.
echo  ============================================================
echo       superAIAutoCutVideo Installation Script
echo       China Mirror - No VPN Required
echo  ============================================================
echo.

:: China mirror configuration
set PIP_MIRROR=https://pypi.tuna.tsinghua.edu.cn/simple
set NPM_MIRROR=https://registry.npmmirror.com

echo   Mirrors:
echo     pip: Tsinghua University
echo     npm: Taobao
echo     model: ModelScope (Aliyun)
echo.

:: ========== Check Python ==========
echo [Step 0] Checking environment...
echo.

echo   Checking Python...
set PYTHON_OK=0
where python >nul 2>&1
if errorlevel 1 goto :no_python

:: Verify it's real Python
python -c "print('ok')" >nul 2>&1
if errorlevel 1 goto :no_python

for /f "tokens=*" %%i in ('python --version 2^>^&1') do set PYVER=%%i
echo   [OK] !PYVER!
set PYTHON_OK=1
goto :check_node

:no_python
echo   [X] Python not found
echo.
echo   ============================================
echo   Please install Python first:
echo.
echo   China Mirror (Fast):
echo   https://mirrors.huaweicloud.com/python/3.11.9/python-3.11.9-amd64.exe
echo.
echo   [IMPORTANT] Check: Add Python to PATH
echo   ============================================
echo.
echo   After installing Python, run this script again.
echo.
goto :end

:check_node
:: ========== Check Node.js ==========
echo   Checking Node.js...
set NODE_OK=0
where node >nul 2>&1
if errorlevel 1 goto :no_node

for /f "tokens=*" %%i in ('node --version 2^>^&1') do set NODEVER=%%i
echo   [OK] Node.js !NODEVER!
set NODE_OK=1
goto :check_git

:no_node
echo   [X] Node.js not found
echo.
echo   ============================================
echo   Please install Node.js first:
echo.
echo   China Mirror (Fast):
echo   https://mirrors.huaweicloud.com/nodejs/v20.18.0/node-v20.18.0-x64.msi
echo.
echo   Official: https://nodejs.org/
echo   ============================================
echo.
echo   After installing Node.js, run this script again.
echo.
goto :end

:check_git
:: ========== Check Git (optional) ==========
echo   Checking Git...
set GIT_OK=0
where git >nul 2>&1
if errorlevel 1 (
    echo   [!] Git not found (optional)
    echo       IndexTTS2 requires Git, but you can skip it.
    echo.
    echo   China Mirror:
    echo   https://mirrors.huaweicloud.com/git-for-windows/v2.47.0.windows.2/Git-2.47.0.2-64-bit.exe
    echo.
) else (
    echo   [OK] Git available
    set GIT_OK=1
)

echo.
echo   Environment check passed!
echo.
echo ============================================================
echo.

:: ========== Step 1: Create venv ==========
echo [1/6] Creating Python virtual environment...

:: Navigate to backend directory
cd /d "%~dp0"
cd ..
cd backend
if errorlevel 1 (
    echo   [X] Cannot find backend directory
    goto :end
)

if exist "venv\Scripts\activate.bat" (
    echo       Already exists, skipping
) else (
    echo       Creating...
    python -m venv venv
    if errorlevel 1 (
        echo       [X] Failed to create venv
        echo       Please reinstall Python with "Add to PATH" checked
        goto :end
    )
    echo       [OK] Created
)
echo.

:: ========== Step 2: Configure pip ==========
echo [2/6] Configuring pip mirror...
call venv\Scripts\activate.bat
if errorlevel 1 (
    echo       [X] Failed to activate venv
    goto :end
)
python -m pip config set global.index-url %PIP_MIRROR% >nul 2>&1
python -m pip install --upgrade pip -i %PIP_MIRROR% -q
echo       [OK] Done
echo.

:: ========== Step 3: Install backend ==========
echo [3/6] Installing Python packages...
echo       This takes 2-5 minutes, please wait...
pip install -r requirements.txt -i %PIP_MIRROR%
if errorlevel 1 (
    echo       [!] Some packages failed
) else (
    echo       [OK] Done
)
echo.

:: ========== Step 4: Install IndexTTS2 ==========
echo [4/6] Installing IndexTTS2...

if "!GIT_OK!"=="0" (
    echo       [SKIP] Git not installed
    goto :step5
)

echo       Trying Gitee mirror...
pip install git+https://gitee.com/mirrors/index-tts.git -i %PIP_MIRROR% 2>nul
if errorlevel 1 (
    echo       Trying GitHub...
    pip install git+https://github.com/index-tts/index-tts.git -i %PIP_MIRROR% 2>nul
    if errorlevel 1 (
        echo       [!] Failed - install manually later
    ) else (
        echo       [OK] Done
    )
) else (
    echo       [OK] Done
)

:step5
echo.

:: ========== Step 5: Install frontend ==========
echo [5/6] Installing frontend packages...

cd /d "%~dp0"
cd ..
cd frontend
if errorlevel 1 (
    echo       [X] Cannot find frontend directory
    goto :end
)

call npm config set registry %NPM_MIRROR%
echo       This takes 1-3 minutes, please wait...
call npm install
if errorlevel 1 (
    echo       [!] Failed
) else (
    echo       [OK] Done
)
echo.

:: ========== Step 6: Download model ==========
echo [6/6] Downloading IndexTTS2 model...

cd /d "%~dp0"
cd ..
cd backend
call venv\Scripts\activate.bat

if exist "checkpoints\config.yaml" (
    echo       Already exists, skipping
    goto :done
)

echo       Model size: 3-5GB
echo       Downloading from ModelScope (China server)...
echo       This takes 10-30 minutes...
echo.

pip install modelscope -i %PIP_MIRROR% -q
python -c "from modelscope import snapshot_download; snapshot_download('IndexTeam/IndexTTS-1.5', local_dir='./checkpoints')"
if errorlevel 1 (
    echo.
    echo       [!] Download failed
    echo.
    echo       Manual download:
    echo       https://modelscope.cn/models/IndexTeam/IndexTTS-1.5
    echo       Put files in: backend\checkpoints\
) else (
    echo       [OK] Done
)

:done
echo.
color 0A
echo  ============================================================
echo                    Installation Complete!
echo  ============================================================
echo.
echo  To start: Double-click scripts\start.bat
echo.
echo  For IndexTTS2, put 6 voice WAV files in:
echo    backend\serviceData\index_tts\voices\
echo.
echo ============================================================

:end
echo.
echo Press any key to exit...
pause >nul
endlocal
