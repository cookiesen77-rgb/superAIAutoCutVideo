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

echo   Installing missing packages...
echo.

pip install uvicorn -i https://pypi.tuna.tsinghua.edu.cn/simple
pip install fastapi -i https://pypi.tuna.tsinghua.edu.cn/simple
pip install pydantic -i https://pypi.tuna.tsinghua.edu.cn/simple
pip install aiohttp -i https://pypi.tuna.tsinghua.edu.cn/simple
pip install httpx -i https://pypi.tuna.tsinghua.edu.cn/simple

echo.
echo   [OK] Done!
echo.
echo   Now try running start.bat again.
echo.
pause

