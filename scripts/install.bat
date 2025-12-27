@echo off

:: Prevent flash close
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

:: China mirror
set PIP_MIRROR=https://pypi.tuna.tsinghua.edu.cn/simple
set NPM_MIRROR=https://registry.npmmirror.com

echo   Mirrors: Tsinghua (pip) / Taobao (npm) / ModelScope (model)
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
    python -c "print('test')" >nul 2>&1
    if not errorlevel 1 (
        set PYTHON_CMD=python
        set PYTHON_OK=1
        for /f "tokens=*" %%i in ('python --version 2^>^&1') do echo   [OK] %%i ^(in PATH^)
        goto :found_python
    )
)

:: Method 2: Search common locations
echo   Not in PATH, searching installation folders...

:: Check AppData Local (most common for user install)
for /d %%i in ("%LOCALAPPDATA%\Programs\Python\Python3*") do (
    if exist "%%i\python.exe" (
        set PYTHON_CMD=%%i\python.exe
        set PYTHON_OK=1
        echo   [OK] Found: %%i
        goto :found_python
    )
)

:: Check C:\Python3xx
for /d %%i in ("C:\Python3*") do (
    if exist "%%i\python.exe" (
        set PYTHON_CMD=%%i\python.exe
        set PYTHON_OK=1
        echo   [OK] Found: %%i
        goto :found_python
    )
)

:: Check Program Files
for /d %%i in ("%ProgramFiles%\Python3*") do (
    if exist "%%i\python.exe" (
        set PYTHON_CMD=%%i\python.exe
        set PYTHON_OK=1
        echo   [OK] Found: %%i
        goto :found_python
    )
)

:: Not found
echo   [X] Python not found
echo.
echo   ============================================
echo   Please install Python:
echo.
echo   Download (China Mirror):
echo   https://mirrors.huaweicloud.com/python/3.11.9/python-3.11.9-amd64.exe
echo.
echo   [IMPORTANT] During install, CHECK:
echo   [x] Add Python to PATH
echo.
echo   Then restart this script.
echo   ============================================
goto :end

:found_python

:: ========== Find Node.js ==========
echo.
echo   Looking for Node.js...

set NODE_OK=0
where node >nul 2>&1
if not errorlevel 1 (
    for /f "tokens=*" %%i in ('node --version 2^>^&1') do echo   [OK] Node.js %%i
    set NODE_OK=1
) else (
    echo   [X] Node.js not found
    echo.
    echo   ============================================
    echo   Please install Node.js:
    echo.
    echo   Download (China Mirror):
    echo   https://mirrors.huaweicloud.com/nodejs/v20.18.0/node-v20.18.0-x64.msi
    echo.
    echo   Official: https://nodejs.org/
    echo   ============================================
    goto :end
)

:: ========== Check Git ==========
echo.
echo   Looking for Git...
set GIT_OK=0
where git >nul 2>&1
if not errorlevel 1 (
    echo   [OK] Git available
    set GIT_OK=1
) else (
    echo   [!] Git not found (optional for IndexTTS2)
)

echo.
echo   Environment OK!
echo.
echo ============================================================
echo.

:: ========== Step 1: Create venv ==========
echo [1/6] Creating virtual environment...

cd /d "%~dp0.."
cd backend
if not exist "backend" if not exist "modules" (
    echo   [X] Wrong directory
    goto :end
)

if exist "venv\Scripts\activate.bat" (
    echo       Already exists
) else (
    echo       Creating with: !PYTHON_CMD!
    "!PYTHON_CMD!" -m venv venv
    if errorlevel 1 (
        echo       [X] Failed
        goto :end
    )
    echo       [OK] Created
)
echo.

:: ========== Step 2: pip config ==========
echo [2/6] Configuring pip...
call venv\Scripts\activate.bat
python -m pip config set global.index-url %PIP_MIRROR% >nul 2>&1
python -m pip install --upgrade pip -i %PIP_MIRROR% -q
echo       [OK] Using Tsinghua mirror
echo.

:: ========== Step 3: Backend packages ==========
echo [3/6] Installing Python packages (2-5 min)...
pip install -r requirements.txt -i %PIP_MIRROR%
if errorlevel 1 (
    echo       [!] Some failed
) else (
    echo       [OK] Done
)
echo.

:: ========== Step 4: IndexTTS2 ==========
echo [4/6] Installing IndexTTS2...
if "!GIT_OK!"=="0" (
    echo       [SKIP] No Git
    goto :step5
)

pip install git+https://gitee.com/mirrors/index-tts.git -i %PIP_MIRROR% 2>nul
if errorlevel 1 (
    pip install git+https://github.com/index-tts/index-tts.git -i %PIP_MIRROR% 2>nul
    if errorlevel 1 (
        echo       [!] Failed
    ) else (
        echo       [OK] Done
    )
) else (
    echo       [OK] Done
)

:step5
echo.

:: ========== Step 5: Frontend ==========
echo [5/6] Installing frontend packages (1-3 min)...
cd /d "%~dp0.."
cd frontend
call npm config set registry %NPM_MIRROR%
call npm install
if errorlevel 1 (
    echo       [!] Failed
) else (
    echo       [OK] Done
)
echo.

:: ========== Step 6: Model ==========
echo [6/6] Downloading model...
cd /d "%~dp0.."
cd backend
call venv\Scripts\activate.bat

if exist "checkpoints\config.yaml" (
    echo       Already exists
    goto :done
)

echo       Size: 3-5GB from ModelScope
echo       Time: 10-30 minutes
echo.
pip install modelscope -i %PIP_MIRROR% -q
python -c "from modelscope import snapshot_download; snapshot_download('IndexTeam/IndexTTS-1.5', local_dir='./checkpoints')"
if errorlevel 1 (
    echo       [!] Failed - download manually:
    echo       https://modelscope.cn/models/IndexTeam/IndexTTS-1.5
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
echo  Start: Double-click scripts\start.bat
echo.
echo  Voice files for IndexTTS2:
echo    Put 6 WAV files in: backend\serviceData\index_tts\voices\
echo.
echo ============================================================

:end
echo.
pause
endlocal
