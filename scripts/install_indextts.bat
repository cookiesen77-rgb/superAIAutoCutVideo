@echo off

:: Prevent flash close
if "%~1"=="" (
    start cmd /k "%~f0" run
    exit /b
)

setlocal enabledelayedexpansion
chcp 65001 >nul 2>&1
title Install IndexTTS
color 0E

echo.
echo  ============================================================
echo       Install IndexTTS from Local Folder
echo       No Git Required
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

:: Check if already installed
pip show indextts >nul 2>&1
if not errorlevel 1 (
    echo   IndexTTS is already installed!
    pip show indextts | findstr "Version"
    echo.
    echo   To reinstall, first run:
    echo   pip uninstall indextts -y
    goto :end
)

echo   Searching for IndexTTS folder...
echo.

set INDEXTTS_PATH=

:: List of paths to check
for %%p in (
    "%USERPROFILE%\Desktop\index-tts-main"
    "%USERPROFILE%\Desktop\index-tts"
    "%USERPROFILE%\Desktop\IndexTTS-main"
    "%USERPROFILE%\Desktop\IndexTTS"
    "%USERPROFILE%\Downloads\index-tts-main"
    "%USERPROFILE%\Downloads\index-tts"
    "%USERPROFILE%\Downloads\IndexTTS-main"
) do (
    if exist "%%~p\setup.py" (
        set "INDEXTTS_PATH=%%~p"
        goto :found
    )
    if exist "%%~p\pyproject.toml" (
        set "INDEXTTS_PATH=%%~p"
        goto :found
    )
)

:: Not found
echo   [X] IndexTTS folder not found
echo.
echo   ============================================
echo   Please:
echo   1. Download: https://github.com/index-tts/index-tts/archive/refs/heads/main.zip
echo   2. Extract ZIP to Desktop
echo   3. Run this script again
echo   ============================================
echo.
echo   Or install manually:
echo   pip install "path\to\index-tts-main" --no-deps -i https://pypi.tuna.tsinghua.edu.cn/simple
goto :end

:found
echo   [OK] Found: !INDEXTTS_PATH!
echo.
echo   Installing IndexTTS...
echo   This may take 1-2 minutes...
echo.

:: Install with --no-deps to avoid version conflicts
pip install "!INDEXTTS_PATH!" --no-deps -i https://pypi.tuna.tsinghua.edu.cn/simple

if errorlevel 1 (
    echo.
    echo   [X] Installation failed
    echo.
    echo   Try manual install:
    echo   cd !INDEXTTS_PATH!
    echo   pip install . --no-deps
    goto :end
)

echo.
echo   Installing additional dependencies...
pip install WeTextProcessing -i https://pypi.tuna.tsinghua.edu.cn/simple 2>nul

echo.
echo   [OK] IndexTTS installed successfully!
echo.
echo   Next: Run scripts\download_model.bat to get the model

:end
echo.
echo Press any key to exit...
pause >nul
endlocal
