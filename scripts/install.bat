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

echo   Mirrors: Tsinghua / Taobao / ModelScope
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
        python --version
        goto :found_python
    )
)

:: Method 2: Search common locations
echo   Not in PATH, searching...

for /d %%i in ("%LOCALAPPDATA%\Programs\Python\Python3*") do (
    if exist "%%i\python.exe" (
        set "PYTHON_CMD=%%i\python.exe"
        set PYTHON_OK=1
        echo   [OK] Found: %%i
        goto :found_python
    )
)

for /d %%i in ("C:\Python3*") do (
    if exist "%%i\python.exe" (
        set "PYTHON_CMD=%%i\python.exe"
        set PYTHON_OK=1
        echo   [OK] Found: %%i
        goto :found_python
    )
)

for /d %%i in ("%ProgramFiles%\Python3*") do (
    if exist "%%i\python.exe" (
        set "PYTHON_CMD=%%i\python.exe"
        set PYTHON_OK=1
        echo   [OK] Found: %%i
        goto :found_python
    )
)

echo   [X] Python not found
echo.
echo   Download: https://mirrors.huaweicloud.com/python/3.11.9/python-3.11.9-amd64.exe
echo   Check: Add Python to PATH
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
goto :check_git

:no_node
echo   [X] Node.js not found
echo.
echo   Download: https://mirrors.huaweicloud.com/nodejs/v20.18.0/node-v20.18.0-x64.msi
goto :end

:check_git
:: ========== Check Git ==========
echo.
echo   Looking for Git...
set GIT_OK=0
where git >nul 2>&1
if not errorlevel 1 (
    echo   [OK] Git available
    set GIT_OK=1
) else (
    echo   [!] Git not found - optional
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

if exist "venv\Scripts\activate.bat" (
    echo       Already exists
) else (
    echo       Creating...
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
echo       [OK] Done
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

echo       Size: 3-5GB
echo       Time: 10-30 min
echo.
pip install modelscope -i %PIP_MIRROR% -q
python -c "from modelscope import snapshot_download; snapshot_download('IndexTeam/IndexTTS-1.5', local_dir='./checkpoints')"
if errorlevel 1 (
    echo       [!] Failed
    echo       Manual: https://modelscope.cn/models/IndexTeam/IndexTTS-1.5
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
echo ============================================================

:end
echo.
pause
endlocal
