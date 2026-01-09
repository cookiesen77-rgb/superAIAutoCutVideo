@echo off

:: Prevent flash close
if "%~1"=="" (
    cmd /k "%~f0" run
    exit /b
)

setlocal enabledelayedexpansion
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
    echo   [X] Virtual environment not found!
    echo   Please run install.bat first.
    goto :end
)

call venv\Scripts\activate.bat

echo   Using mirror: Tsinghua University
echo.
set PIP_MIRROR=https://pypi.tuna.tsinghua.edu.cn/simple

echo   [1/7] Core packages...
pip install fastapi uvicorn[standard] pydantic httpx -i %PIP_MIRROR% -q
echo         Done

echo   [2/7] Media packages...
pip install opencv-python ffmpeg-python Pillow -i %PIP_MIRROR% -q
echo         Done

echo   [3/7] PyTorch (this may take a while)...
pip install torch torchvision numpy -i %PIP_MIRROR% -q
echo         Done

echo   [4/7] Web packages...
pip install websockets python-multipart aiohttp requests -i %PIP_MIRROR% -q
echo         Done

echo   [5/7] TTS packages...
pip install edge-tts pypinyin cn2an pyyaml soundfile librosa -i %PIP_MIRROR% -q
echo         Done

echo   [6/7] AI packages...
pip install transformers accelerate modelscope huggingface_hub -i %PIP_MIRROR% -q
echo         Done

echo   [7/7] Other packages...
pip install python-dotenv loguru psutil tencentcloud-sdk-python -i %PIP_MIRROR% -q
echo         Done

echo.
echo  ============================================================
echo                    Packages Fixed!
echo  ============================================================
echo.
echo   If IndexTTS2 is still missing, run:
echo   scripts\install_indextts.bat
echo.

:end
echo.
pause
endlocal
