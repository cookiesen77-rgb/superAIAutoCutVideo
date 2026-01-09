@echo off

:: Prevent flash close - run in new window
if "%~1"=="" (
    start cmd /k "%~f0" run
    exit /b
)

setlocal enabledelayedexpansion
chcp 65001 >nul 2>&1
title superAIAutoCutVideo Setup
color 0A

echo.
echo  ============================================================
echo       superAIAutoCutVideo Installation Script
echo       China Mirror - No VPN Required - No Git Required
echo  ============================================================
echo.

:: China mirror
set PIP_MIRROR=https://pypi.tuna.tsinghua.edu.cn/simple
set NPM_MIRROR=https://registry.npmmirror.com

echo   Mirrors: Tsinghua (pip) / Taobao (npm)
echo.

:: ========== Find Python ==========
echo [Step 0] Checking environment...
echo.
echo   Looking for Python...

set PYTHON_CMD=
set PYTHON_OK=0

:: Method 1: Check PATH
where python >nul 2>&1
if not errorlevel 1 (
    for /f "tokens=*" %%i in ('python --version 2^>^&1') do set PYVER=%%i
    echo !PYVER! | findstr /C:"3.14" >nul
    if not errorlevel 1 (
        echo   [!] Python 3.14 detected - NOT compatible!
        echo   Please install Python 3.11 instead.
        goto :python_error
    )
    python -c "print('test')" >nul 2>&1
    if not errorlevel 1 (
        set PYTHON_CMD=python
        set PYTHON_OK=1
        echo   [OK] !PYVER!
        goto :found_python
    )
)

:: Method 2: Search common locations
echo   Not in PATH, searching...

for /d %%i in ("%LOCALAPPDATA%\Programs\Python\Python311*") do (
    if exist "%%i\python.exe" (
        set "PYTHON_CMD=%%i\python.exe"
        set PYTHON_OK=1
        echo   [OK] Found: %%i
        goto :found_python
    )
)

for /d %%i in ("%LOCALAPPDATA%\Programs\Python\Python310*") do (
    if exist "%%i\python.exe" (
        set "PYTHON_CMD=%%i\python.exe"
        set PYTHON_OK=1
        echo   [OK] Found: %%i
        goto :found_python
    )
)

for /d %%i in ("C:\Python311*") do (
    if exist "%%i\python.exe" (
        set "PYTHON_CMD=%%i\python.exe"
        set PYTHON_OK=1
        echo   [OK] Found: %%i
        goto :found_python
    )
)

:python_error
echo   [X] Python 3.11 not found
echo.
echo   ============================================
echo   Please install Python 3.11 (NOT 3.14!)
echo   Download: https://mirrors.huaweicloud.com/python/3.11.9/python-3.11.9-amd64.exe
echo   [IMPORTANT] Check: Add Python to PATH
echo   ============================================
goto :end

:found_python

:: ========== Find Node.js ==========
echo.
echo   Looking for Node.js...

set NODE_OK=0
where node >nul 2>&1
if errorlevel 1 goto :no_node

node --version >nul 2>&1
if errorlevel 1 goto :no_node

echo   [OK] Node.js found
node --version
set NODE_OK=1
goto :env_ok

:no_node
echo   [X] Node.js not found
echo.
echo   Download: https://mirrors.huaweicloud.com/nodejs/v20.18.0/node-v20.18.0-x64.msi
goto :end

:env_ok
echo.
echo   Environment OK!
echo.
echo ============================================================
echo.

:: ========== Step 1: Create venv ==========
echo [1/5] Creating virtual environment...

cd /d "%~dp0.."
cd backend

if exist "venv\Scripts\activate.bat" (
    echo       Already exists, skipping
) else (
    echo       Creating new venv...
    "!PYTHON_CMD!" -m venv venv
    if errorlevel 1 (
        echo       [X] Failed to create venv
        goto :end
    )
    echo       [OK] Created
)
echo.

:: ========== Step 2: pip config ==========
echo [2/5] Configuring pip...
call venv\Scripts\activate.bat
python -m pip config set global.index-url %PIP_MIRROR% >nul 2>&1
python -m pip install --upgrade pip -i %PIP_MIRROR% -q
echo       [OK] Done
echo.

:: ========== Step 3: Backend packages ==========
echo [3/5] Installing Python packages (5-10 min)...
echo       Please wait...
pip install -r requirements.txt -i %PIP_MIRROR%
if errorlevel 1 (
    echo.
    echo       [!] Some packages failed, trying core packages...
    pip install fastapi uvicorn[standard] pydantic httpx websockets -i %PIP_MIRROR%
    pip install opencv-python numpy==1.26.2 Pillow -i %PIP_MIRROR%
    pip install torch==2.5.1 torchvision==0.20.1 torchaudio==2.5.1 -i %PIP_MIRROR%
    pip install edge-tts pypinyin cn2an pyyaml soundfile librosa -i %PIP_MIRROR%
    pip install transformers accelerate requests aiohttp -i %PIP_MIRROR%
    pip install python-multipart python-dotenv loguru psutil -i %PIP_MIRROR%
)
echo       [OK] Backend packages done
echo.

:: ========== Step 4: IndexTTS ==========
echo [4/5] Installing IndexTTS...
echo.

:: Check if already installed
pip show indextts >nul 2>&1
if not errorlevel 1 (
    echo       Already installed
    goto :step5
)

:: Search for local folder
set INDEXTTS_LOCAL=

:: Check Desktop
for %%p in (
    "%USERPROFILE%\Desktop\index-tts-main"
    "%USERPROFILE%\Desktop\index-tts"
    "%USERPROFILE%\Desktop\IndexTTS-main"
    "%USERPROFILE%\Desktop\IndexTTS"
) do (
    if exist "%%~p\setup.py" (
        set "INDEXTTS_LOCAL=%%~p"
        goto :found_indextts
    )
)

:: Check Downloads
for %%p in (
    "%USERPROFILE%\Downloads\index-tts-main"
    "%USERPROFILE%\Downloads\index-tts"
) do (
    if exist "%%~p\setup.py" (
        set "INDEXTTS_LOCAL=%%~p"
        goto :found_indextts
    )
)

:: Not found
echo       [!] IndexTTS folder not found
echo.
echo       Please download and extract to Desktop:
echo       https://github.com/index-tts/index-tts/archive/refs/heads/main.zip
echo.
echo       Then run: scripts\install_indextts.bat
goto :step5

:found_indextts
echo       Found: !INDEXTTS_LOCAL!
echo       Installing (this takes 1-2 min)...
pip install "!INDEXTTS_LOCAL!" --no-deps -i %PIP_MIRROR%
if errorlevel 1 (
    echo       [!] Install failed
) else (
    echo       [OK] IndexTTS installed
    :: Install IndexTTS dependencies
    pip install WeTextProcessing -i %PIP_MIRROR% 2>nul
)

:step5
echo.

:: ========== Step 5: Frontend ==========
echo [5/5] Installing frontend packages (1-3 min)...
cd /d "%~dp0.."
cd frontend
call npm config set registry %NPM_MIRROR%
call npm install
if errorlevel 1 (
    echo       [!] npm install failed
) else (
    echo       [OK] Frontend done
)
echo.

:: ========== Check Model ==========
echo ============================================================
echo.
echo   Checking model files...
cd /d "%~dp0.."
cd backend

if exist "checkpoints\gpt.pth" (
    echo   [OK] Model files found
) else (
    echo   [!] Model not found
    echo.
    echo   Please download model:
    echo   1. Run: scripts\download_model.bat
    echo   OR
    echo   2. Manual: https://hf-mirror.com/IndexTeam/Index-TTS/tree/main
    echo      Put files in: backend\checkpoints\
)

echo.
color 0A
echo  ============================================================
echo                    Installation Complete!
echo  ============================================================
echo.
echo  Next steps:
echo  1. If IndexTTS not installed: scripts\install_indextts.bat
echo  2. If model not found: scripts\download_model.bat
echo  3. Start app: scripts\start.bat
echo.
echo ============================================================

:end
echo.
echo Press any key to exit...
pause >nul
endlocal
