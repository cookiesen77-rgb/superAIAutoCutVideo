@echo off
chcp 65001 >nul
echo ============================================================
echo     superAIAutoCutVideo 安装脚本 (Windows)
echo ============================================================
echo.

:: 检查 Python
python --version >nul 2>&1
if errorlevel 1 (
    echo [错误] 未检测到 Python，请先安装 Python 3.10+
    echo 下载地址: https://www.python.org/downloads/
    pause
    exit /b 1
)

:: 检查 Node.js
node --version >nul 2>&1
if errorlevel 1 (
    echo [错误] 未检测到 Node.js，请先安装 Node.js 18+
    echo 下载地址: https://nodejs.org/
    pause
    exit /b 1
)

echo [1/5] 创建 Python 虚拟环境...
cd /d "%~dp0..\backend"
if exist "venv" (
    echo      虚拟环境已存在，跳过创建
) else (
    python -m venv venv
    echo      虚拟环境创建成功
)

echo.
echo [2/5] 激活虚拟环境并安装后端依赖...
call venv\Scripts\activate.bat
python -m pip install --upgrade pip -q
pip install -r requirements.txt -q
echo      后端依赖安装完成

echo.
echo [3/5] 安装 IndexTTS2...
pip install git+https://github.com/index-tts/index-tts.git -q 2>nul
if errorlevel 1 (
    echo      [警告] IndexTTS2 安装失败，请稍后手动安装
) else (
    echo      IndexTTS2 安装完成
)

echo.
echo [4/5] 安装前端依赖...
cd /d "%~dp0..\frontend"
call npm install --silent
echo      前端依赖安装完成

echo.
echo [5/5] 下载 IndexTTS2 模型...
cd /d "%~dp0..\backend"
call venv\Scripts\activate.bat
python -c "from modelscope import snapshot_download; snapshot_download('IndexTeam/IndexTTS-1.5', local_dir='./checkpoints')" 2>nul
if errorlevel 1 (
    echo      [警告] 模型下载失败，请稍后手动下载
    echo      参考文档: docs/IndexTTS2_使用指南.md
) else (
    echo      模型下载完成
)

echo.
echo ============================================================
echo     安装完成！
echo ============================================================
echo.
echo 启动方式:
echo   方式1: 双击 scripts\start.bat
echo   方式2: 手动启动（见下方命令）
echo.
echo 手动启动后端:
echo   cd backend
echo   venv\Scripts\activate
echo   python main.py
echo.
echo 手动启动前端:
echo   cd frontend
echo   npm run dev
echo.
echo 注意: 首次使用需准备音色文件，详见使用指南
echo ============================================================
pause

