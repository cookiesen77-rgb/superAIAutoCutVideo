@echo off
chcp 65001 >nul 2>&1
title Fix Missing Packages
color 0E

echo.
echo  ============================================================
echo       Fix Missing Python Packages
echo  ============================================================
echo.

cd /d "%~dp0.."
cd backend

if not exist "venv\Scripts\activate.bat" (
    echo   [X] Virtual environment not found
    echo   Please run install.bat first
    pause
    exit /b 1
)

call venv\Scripts\activate.bat

:: Check Python version
python --version
echo.

echo   Installing core packages (pre-built wheels)...
echo.

:: Install packages one by one with specific versions that have wheels
pip install --only-binary :all: numpy -i https://pypi.tuna.tsinghua.edu.cn/simple 2>nul
if errorlevel 1 (
    echo   [!] numpy wheel not available for your Python version
    echo   Trying older version...
    pip install "numpy<2.0" -i https://pypi.tuna.tsinghua.edu.cn/simple
)

pip install uvicorn -i https://pypi.tuna.tsinghua.edu.cn/simple
pip install fastapi -i https://pypi.tuna.tsinghua.edu.cn/simple
pip install pydantic -i https://pypi.tuna.tsinghua.edu.cn/simple
pip install aiohttp -i https://pypi.tuna.tsinghua.edu.cn/simple
pip install httpx -i https://pypi.tuna.tsinghua.edu.cn/simple
pip install python-multipart -i https://pypi.tuna.tsinghua.edu.cn/simple
pip install aiofiles -i https://pypi.tuna.tsinghua.edu.cn/simple
pip install requests -i https://pypi.tuna.tsinghua.edu.cn/simple
pip install edge-tts -i https://pypi.tuna.tsinghua.edu.cn/simple

echo.
echo   ============================================================
echo.

:: Check if numpy failed
python -c "import numpy" 2>nul
if errorlevel 1 (
    color 0C
    echo   [X] numpy installation failed
    echo.
    echo   Your Python version (3.14) is too new!
    echo   numpy does not have pre-built packages for Python 3.14
    echo.
    echo   SOLUTION:
    echo   1. Uninstall Python 3.14
    echo   2. Install Python 3.11:
    echo      https://mirrors.huaweicloud.com/python/3.11.9/python-3.11.9-amd64.exe
    echo   3. Delete backend\venv folder
    echo   4. Run install.bat again
    echo.
) else (
    color 0A
    echo   [OK] All packages installed!
    echo.
    echo   Now try running start.bat
    echo.
)

pause
