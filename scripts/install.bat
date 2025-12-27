@echo off
chcp 65001 >nul
title superAIAutoCutVideo 安装程序
color 0A

echo.
echo  ╔════════════════════════════════════════════════════════════╗
echo  ║      superAIAutoCutVideo 一键安装脚本 (Windows)            ║
echo  ║      支持虚拟环境 - 使用国内镜像源                         ║
echo  ╚════════════════════════════════════════════════════════════╝
echo.

:: 国内镜像配置
set PIP_MIRROR=https://pypi.tuna.tsinghua.edu.cn/simple
set NPM_MIRROR=https://registry.npmmirror.com

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

echo   [√] 使用国内镜像源（无需 VPN）
echo       pip: %PIP_MIRROR%
echo       npm: %NPM_MIRROR%

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

:: ========== 步骤 2: 配置 pip 国内镜像并升级 ==========
echo [2/6] 配置 pip 国内镜像并升级...
call venv\Scripts\activate.bat
python -m pip config set global.index-url %PIP_MIRROR% >nul 2>&1
python -m pip install --upgrade pip -i %PIP_MIRROR% --quiet
echo       [√] pip 已配置清华镜像
echo.

:: ========== 步骤 3: 安装后端依赖 ==========
echo [3/6] 安装后端 Python 依赖...
echo       这可能需要几分钟，请耐心等待...
pip install -r requirements.txt -i %PIP_MIRROR%
if errorlevel 1 (
    color 0E
    echo       [!] 部分依赖安装失败，请检查网络
) else (
    echo       [√] 后端依赖安装完成
)
echo.

:: ========== 步骤 4: 安装 IndexTTS2 ==========
echo [4/6] 安装 IndexTTS2...
echo       正在从 Gitee 镜像安装（国内加速）...

:: 尝试从 Gitee 镜像安装
pip install git+https://gitee.com/mirrors/index-tts.git -i %PIP_MIRROR% 2>nul
if errorlevel 1 (
    echo       Gitee 镜像失败，尝试 GitHub...
    pip install git+https://github.com/index-tts/index-tts.git -i %PIP_MIRROR% 2>nul
    if errorlevel 1 (
        color 0E
        echo       [!] IndexTTS2 安装失败
        echo.
        echo       请手动安装，方法：
        echo       1. 下载: https://github.com/index-tts/index-tts/archive/refs/heads/main.zip
        echo       2. 解压后进入目录，运行: pip install .
        echo.
    ) else (
        echo       [√] IndexTTS2 安装完成（GitHub）
    )
) else (
    echo       [√] IndexTTS2 安装完成（Gitee 镜像）
)
echo.

:: ========== 步骤 5: 安装前端依赖 ==========
echo [5/6] 安装前端 npm 依赖...
cd /d "%~dp0..\frontend"
echo       配置 npm 淘宝镜像...
call npm config set registry %NPM_MIRROR%
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
echo       正在从 ModelScope 下载（国内服务器，无需 VPN）...
echo.

:: 先安装 modelscope
pip install modelscope -i %PIP_MIRROR% --quiet

python -c "from modelscope import snapshot_download; snapshot_download('IndexTeam/IndexTTS-1.5', local_dir='./checkpoints')"
if errorlevel 1 (
    color 0E
    echo.
    echo       [!] 模型下载失败
    echo.
    echo       请手动下载，方法：
    echo       1. 访问: https://modelscope.cn/models/IndexTeam/IndexTTS-1.5
    echo       2. 点击"下载模型"，下载所有文件
    echo       3. 解压到 backend\checkpoints\ 目录
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
echo  使用的国内镜像:
echo    pip:  清华大学 (%PIP_MIRROR%)
echo    npm:  淘宝镜像 (%NPM_MIRROR%)
echo    模型: ModelScope（阿里云）
echo.
echo ============================================================
pause
