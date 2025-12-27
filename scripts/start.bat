@echo off
chcp 65001 >nul 2>&1
title superAIAutoCutVideo
color 0A

echo  ============================================================
echo       superAIAutoCutVideo - Starting Services
echo  ============================================================
echo.

:: Start backend
echo [Starting Backend Server...]
cd /d "%~dp0..\backend"
if not exist "venv\Scripts\activate.bat" (
    color 0C
    echo [ERROR] Virtual environment not found
    echo Please run install.bat first
    pause
    exit /b 1
)

start "Backend Server" cmd /k "call venv\Scripts\activate.bat && python main.py"

:: Wait for backend
echo Waiting for backend to start (5 seconds)...
timeout /t 5 /nobreak >nul

:: Start frontend
echo [Starting Frontend Server...]
cd /d "%~dp0..\frontend"
start "Frontend Server" cmd /k "npm run dev"

echo.
echo  ============================================================
echo       Services Started!
echo  ============================================================
echo.
echo  Backend:  http://localhost:8000
echo  Frontend: http://localhost:5173 (check terminal for actual port)
echo.
echo  To stop: Close both command windows
echo  ============================================================
echo.
pause
