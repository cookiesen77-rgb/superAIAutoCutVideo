@echo off

:: Prevent flash close
if "%~1"=="" (
    start cmd /k "%~f0" run
    exit /b
)

setlocal enabledelayedexpansion
chcp 65001 >nul 2>&1
title IndexTTS Model Download
color 0B

echo.
echo  ============================================================
echo       IndexTTS Model Download
echo       HF-Mirror (China Fast)
echo  ============================================================
echo.

cd /d "%~dp0.."
cd backend

if not exist "venv\Scripts\activate.bat" (
    echo   [X] Virtual environment not found!
    echo   Please run install.bat first.
    goto :end
)

call venv\Scripts\activate.bat

:: Check if model exists
if exist "checkpoints\gpt.pth" (
    echo   [OK] Model already exists in checkpoints\
    echo.
    echo   To re-download, delete the checkpoints folder first.
    goto :end
)

:: Create checkpoints folder
if not exist "checkpoints" mkdir checkpoints

echo   Model size: ~3GB
echo   Estimated time: 10-30 minutes
echo.

:: Install huggingface_hub
pip show huggingface_hub >nul 2>&1
if errorlevel 1 (
    echo   Installing download tools...
    pip install huggingface_hub -i https://pypi.tuna.tsinghua.edu.cn/simple -q
)

:: Set HF Mirror
set HF_ENDPOINT=https://hf-mirror.com

echo   Starting download from hf-mirror.com...
echo.

:: Try download
python -c "import os; os.environ['HF_ENDPOINT']='https://hf-mirror.com'; from huggingface_hub import snapshot_download; snapshot_download('IndexTeam/Index-TTS', local_dir='checkpoints', resume_download=True)"

if exist "checkpoints\gpt.pth" (
    echo.
    echo   ============================================
    echo   [OK] Model download complete!
    echo   ============================================
    echo.
    echo   You can now run: scripts\start.bat
    goto :end
)

:: Download failed
echo.
echo   [X] Automatic download failed
echo.
echo   ============================================
echo   Please download manually:
echo.
echo   1. Open browser: https://hf-mirror.com/IndexTeam/Index-TTS/tree/main
echo.
echo   2. Download these files:
echo      - gpt.pth (main model, ~700MB)
echo      - bigvgan_generator.pth (~500MB)
echo      - bigvgan_discriminator.pth (~1.6GB)
echo      - dvae.pth (~250MB)
echo      - bpe.model
echo      - config.yaml
echo      - unigram_12000.vocab
echo.
echo   3. Put all files in:
echo      %cd%\checkpoints\
echo.
echo   4. Run: scripts\start.bat
echo   ============================================

:end
echo.
echo Press any key to exit...
pause >nul
endlocal
