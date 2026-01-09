@echo off
chcp 65001 >nul 2>&1
setlocal enabledelayedexpansion

echo ============================================================
echo   IndexTTS Dependencies Fix Script
echo ============================================================
echo.

cd /d "%~dp0..\backend"

REM Check if venv exists
if exist "venv\Scripts\activate.bat" (
    echo [INFO] Activating virtual environment...
    call venv\Scripts\activate.bat
) else (
    echo [ERROR] Virtual environment not found!
    echo Please run install.bat first.
    pause
    exit /b 1
)

echo.
echo [1/2] Installing IndexTTS core dependencies...
pip install omegaconf einops sentencepiece munch -i https://pypi.tuna.tsinghua.edu.cn/simple --trusted-host pypi.tuna.tsinghua.edu.cn

echo.
echo [2/2] Installing additional dependencies...
pip install WeTextProcessing jieba g2p-en matplotlib json5 textstat -i https://pypi.tuna.tsinghua.edu.cn/simple --trusted-host pypi.tuna.tsinghua.edu.cn

echo.
echo ============================================================
echo   Installation complete!
echo ============================================================
echo.
echo Please restart the application and try again.
echo.
pause

