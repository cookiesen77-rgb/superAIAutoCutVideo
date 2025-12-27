@echo off
chcp 65001 >nul
echo ============================================================
echo     superAIAutoCutVideo 启动脚本
echo ============================================================
echo.

:: 启动后端
echo [启动后端服务...]
cd /d "%~dp0..\backend"
if not exist "venv\Scripts\activate.bat" (
    echo [错误] 虚拟环境不存在，请先运行 install.bat
    pause
    exit /b 1
)

start "Backend Server" cmd /k "call venv\Scripts\activate.bat && python main.py"

:: 等待后端启动
echo 等待后端启动 (5秒)...
timeout /t 5 /nobreak >nul

:: 启动前端
echo [启动前端服务...]
cd /d "%~dp0..\frontend"
start "Frontend Server" cmd /k "npm run dev"

echo.
echo ============================================================
echo     服务已启动！
echo ============================================================
echo.
echo 后端地址: http://localhost:8000
echo 前端地址: http://localhost:5173 (或查看终端输出)
echo.
echo 关闭方式: 关闭两个命令行窗口即可
echo ============================================================

