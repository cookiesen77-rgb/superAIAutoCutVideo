@echo off
chcp 65001 >nul
title superAIAutoCutVideo 安装程序
color 0A

echo.
echo  ╔════════════════════════════════════════════════════════════╗
echo  ║      superAIAutoCutVideo 一键安装脚本 (Windows)            ║
echo  ║      支持虚拟环境 - 自动下载所有依赖                       ║
echo  ╚════════════════════════════════════════════════════════════╝
echo.

:: ========== 环境检查 ==========
echo [环境检查]
echo.

:: 检查 Python
echo   检查 Python...
python --version >nul 2>&1
if errorlevel 1 (
    color 0C
    echo   [X] 未检测到 Python
    echo.
    echo   请先安装 Python 3.10+
    echo   下载地址: https://www.python.org/downloads/
    echo   安装时请勾选 "Add Python to PATH"
    echo.
    pause
    exit /b 1
)
for /f "tokens=2" %%i in ('python --version 2^>^&1') do set PYVER=%%i
echo   [√] Python %PYVER%

:: 检查 Node.js
echo   检查 Node.js...
node --version >nul 2>&1
if errorlevel 1 (
    color 0C
    echo   [X] 未检测到 Node.js
    echo.
    echo   请先安装 Node.js 18+
    echo   下载地址: https://nodejs.org/
    echo.
    pause
    exit /b 1
)
for /f %%i in ('node --version 2^>^&1') do set NODEVER=%%i
echo   [√] Node.js %NODEVER%

:: 检查 Git（用于安装 IndexTTS）
echo   检查 Git...
git --version >nul 2>&1
if errorlevel 1 (
    echo   [!] 未检测到 Git（IndexTTS2 需要从 GitHub 安装）
    echo       下载地址: https://git-scm.com/downloads
    set GIT_OK=0
) else (
    echo   [√] Git 已安装
    set GIT_OK=1
)

echo.
echo ============================================================
echo.

:: ========== 步骤 1: 创建虚拟环境 ==========
echo [1/6] 创建 Python 虚拟环境...
cd /d "%~dp0..\backend"
if exist "venv\Scripts\activate.bat" (
    echo       虚拟环境已存在，跳过
) else (
    echo       正在创建虚拟环境...
    python -m venv venv
    if errorlevel 1 (
        color 0C
        echo       [X] 虚拟环境创建失败
        pause
        exit /b 1
    )
    echo       [√] 虚拟环境创建成功
)
echo.

:: ========== 步骤 2: 升级 pip ==========
echo [2/6] 升级 pip...
call venv\Scripts\activate.bat
python -m pip install --upgrade pip --quiet
echo       [√] pip 已升级
echo.

:: ========== 步骤 3: 安装后端依赖 ==========
echo [3/6] 安装后端 Python 依赖...
echo       这可能需要几分钟，请耐心等待...
pip install -r requirements.txt
if errorlevel 1 (
    color 0E
    echo       [!] 部分依赖安装失败，请检查网络
) else (
    echo       [√] 后端依赖安装完成
)
echo.

:: ========== 步骤 4: 安装 IndexTTS2 ==========
echo [4/6] 安装 IndexTTS2...
if "%GIT_OK%"=="0" (
    echo       [!] 跳过（需要安装 Git）
) else (
    echo       正在从 GitHub 安装，请等待...
    pip install git+https://github.com/index-tts/index-tts.git
    if errorlevel 1 (
        color 0E
        echo       [!] IndexTTS2 安装失败
        echo       请稍后手动运行: pip install git+https://github.com/index-tts/index-tts.git
    ) else (
        echo       [√] IndexTTS2 安装完成
    )
)
echo.

:: ========== 步骤 5: 安装前端依赖 ==========
echo [5/6] 安装前端 npm 依赖...
cd /d "%~dp0..\frontend"
echo       这可能需要几分钟，请耐心等待...
call npm install
if errorlevel 1 (
    color 0E
    echo       [!] 前端依赖安装失败
) else (
    echo       [√] 前端依赖安装完成
)
echo.

:: ========== 步骤 6: 下载模型 ==========
echo [6/6] 下载 IndexTTS2 模型...
cd /d "%~dp0..\backend"
call venv\Scripts\activate.bat

:: 检查模型是否已存在
if exist "checkpoints\config.yaml" (
    echo       模型已存在，跳过下载
    goto :install_done
)

echo       模型大小约 3-5GB，请确保网络畅通...
echo       正在从 ModelScope 下载...
python -c "from modelscope import snapshot_download; snapshot_download('IndexTeam/IndexTTS-1.5', local_dir='./checkpoints')"
if errorlevel 1 (
    color 0E
    echo.
    echo       [!] 自动下载失败，请手动下载模型
    echo.
    echo       方法1: 使用 ModelScope
    echo         pip install modelscope
    echo         python -c "from modelscope import snapshot_download; snapshot_download('IndexTeam/IndexTTS-1.5', local_dir='./backend/checkpoints')"
    echo.
    echo       方法2: 使用 Hugging Face
    echo         pip install huggingface_hub
    echo         python -c "from huggingface_hub import snapshot_download; snapshot_download('IndexTeam/IndexTTS-1.5', local_dir='./backend/checkpoints')"
    echo.
) else (
    echo       [√] 模型下载完成
)

:install_done
echo.
color 0A
echo  ╔════════════════════════════════════════════════════════════╗
echo  ║                    安装完成！                              ║
echo  ╚════════════════════════════════════════════════════════════╝
echo.
echo  启动方式:
echo    双击 scripts\start.bat
echo.
echo  或手动启动:
echo    后端: cd backend ^&^& venv\Scripts\activate ^&^& python main.py
echo    前端: cd frontend ^&^& npm run dev
echo.
echo  注意事项:
echo    1. 首次使用需准备 6 个音色 WAV 文件
echo       放入: backend\serviceData\index_tts\voices\
echo    2. 详细说明请查看: docs\IndexTTS2_使用指南.md
echo.
echo ============================================================
pause
