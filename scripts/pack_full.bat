@echo off
chcp 65001 >nul
title 打包完整版（含虚拟环境和模型）
color 0E

echo.
echo  ╔════════════════════════════════════════════════════════════╗
echo  ║     打包完整版 - 含虚拟环境和模型                          ║
echo  ║     运行后生成可直接使用的 ZIP 包                          ║
echo  ╚════════════════════════════════════════════════════════════╝
echo.

set PROJECT_DIR=%~dp0..
set PACK_NAME=superAIAutoCutVideo_Full
set PACK_DIR=%USERPROFILE%\Desktop\%PACK_NAME%

:: ========== 步骤 1: 安装依赖 ==========
echo [1/5] 检查并安装依赖...
cd /d "%PROJECT_DIR%"

:: 如果虚拟环境不存在，先运行安装
if not exist "backend\venv\Scripts\activate.bat" (
    echo       虚拟环境不存在，先运行安装脚本...
    call scripts\install.bat
    if errorlevel 1 (
        echo       [X] 安装失败
        pause
        exit /b 1
    )
)
echo       [√] 依赖已安装
echo.

:: ========== 步骤 2: 检查模型 ==========
echo [2/5] 检查模型文件...
if not exist "backend\checkpoints\config.yaml" (
    echo       [!] 模型文件不存在
    echo       请先运行 install.bat 下载模型，或手动下载到 backend\checkpoints\
    pause
    exit /b 1
)
echo       [√] 模型文件已就绪
echo.

:: ========== 步骤 3: 检查音色文件 ==========
echo [3/5] 检查音色文件...
set VOICE_DIR=backend\serviceData\index_tts\voices
set VOICE_COUNT=0
if exist "%VOICE_DIR%\male_youth.wav" set /a VOICE_COUNT+=1
if exist "%VOICE_DIR%\male_mature.wav" set /a VOICE_COUNT+=1
if exist "%VOICE_DIR%\female_sweet.wav" set /a VOICE_COUNT+=1
if exist "%VOICE_DIR%\female_elegant.wav" set /a VOICE_COUNT+=1
if exist "%VOICE_DIR%\female_lively.wav" set /a VOICE_COUNT+=1
if exist "%VOICE_DIR%\child.wav" set /a VOICE_COUNT+=1

if %VOICE_COUNT% LSS 6 (
    echo       [!] 音色文件不完整 (%VOICE_COUNT%/6)
    echo       请将 6 个音色 WAV 文件放入: %VOICE_DIR%\
    echo.
    echo       缺少的文件会导致对应音色不可用，是否继续？
    choice /C YN /M "继续打包"
    if errorlevel 2 exit /b 1
) else (
    echo       [√] 音色文件完整 (6/6)
)
echo.

:: ========== 步骤 4: 创建打包目录 ==========
echo [4/5] 创建打包目录...
if exist "%PACK_DIR%" (
    echo       清理旧打包目录...
    rmdir /s /q "%PACK_DIR%"
)
mkdir "%PACK_DIR%"

echo       复制项目文件...
:: 复制后端（含虚拟环境和模型）
xcopy /E /I /Q "backend" "%PACK_DIR%\backend" >nul
:: 排除 __pycache__
for /d /r "%PACK_DIR%\backend" %%d in (__pycache__) do @if exist "%%d" rmdir /s /q "%%d"

:: 复制前端（不含 node_modules）
mkdir "%PACK_DIR%\frontend"
xcopy /E /I /Q "frontend\src" "%PACK_DIR%\frontend\src" >nul
xcopy /E /I /Q "frontend\public" "%PACK_DIR%\frontend\public" >nul 2>nul
copy "frontend\package.json" "%PACK_DIR%\frontend\" >nul
copy "frontend\package-lock.json" "%PACK_DIR%\frontend\" >nul 2>nul
copy "frontend\vite.config.ts" "%PACK_DIR%\frontend\" >nul 2>nul
copy "frontend\tsconfig.json" "%PACK_DIR%\frontend\" >nul 2>nul
copy "frontend\tsconfig.node.json" "%PACK_DIR%\frontend\" >nul 2>nul
copy "frontend\index.html" "%PACK_DIR%\frontend\" >nul 2>nul
copy "frontend\tailwind.config.js" "%PACK_DIR%\frontend\" >nul 2>nul
copy "frontend\postcss.config.js" "%PACK_DIR%\frontend\" >nul 2>nul

:: 复制脚本和文档
xcopy /E /I /Q "scripts" "%PACK_DIR%\scripts" >nul
xcopy /E /I /Q "docs" "%PACK_DIR%\docs" >nul 2>nul
copy "LICENSE" "%PACK_DIR%\" >nul 2>nul
copy "README.md" "%PACK_DIR%\" >nul 2>nul
copy "快速开始.md" "%PACK_DIR%\" >nul 2>nul

echo       [√] 文件复制完成
echo.

:: ========== 步骤 5: 创建启动脚本 ==========
echo [5/5] 创建快速启动脚本...

:: 创建简化版启动脚本（无需安装，直接运行）
(
echo @echo off
echo chcp 65001 ^>nul
echo title superAIAutoCutVideo
echo color 0A
echo.
echo echo  启动 superAIAutoCutVideo...
echo echo.
echo.
echo :: 启动后端
echo cd /d "%%~dp0backend"
echo start "Backend" cmd /k "call venv\Scripts\activate.bat && python main.py"
echo.
echo :: 等待后端启动
echo timeout /t 5 /nobreak ^>nul
echo.
echo :: 启动前端
echo cd /d "%%~dp0frontend"
echo start "Frontend" cmd /k "npm run dev"
echo.
echo echo.
echo echo  ============================================
echo echo    服务已启动！
echo echo    后端: http://localhost:8000
echo echo    前端: http://localhost:5173
echo echo  ============================================
echo echo.
echo echo  关闭方式: 关闭两个命令行窗口
echo pause
) > "%PACK_DIR%\启动.bat"

:: 创建前端依赖安装脚本（首次使用需要运行一次）
(
echo @echo off
echo chcp 65001 ^>nul
echo title 安装前端依赖
echo echo 正在安装前端依赖...
echo cd /d "%%~dp0frontend"
echo call npm config set registry https://registry.npmmirror.com
echo call npm install
echo echo.
echo echo [√] 前端依赖安装完成！
echo echo 现在可以双击"启动.bat"运行程序了
echo pause
) > "%PACK_DIR%\首次使用-安装前端依赖.bat"

echo       [√] 启动脚本创建完成
echo.

:: ========== 计算大小并显示结果 ==========
echo ============================================================
echo.
echo  打包完成！
echo.
echo  输出目录: %PACK_DIR%
echo.

:: 获取目录大小
for /f "tokens=3" %%a in ('dir /s "%PACK_DIR%" ^| findstr "个文件"') do set SIZE=%%a
echo  预计大小: 约 %SIZE% 字节
echo.
echo  目录结构:
echo    %PACK_NAME%\
echo    ├── backend\          (含 venv 和 checkpoints)
echo    ├── frontend\         (不含 node_modules)
echo    ├── scripts\
echo    ├── docs\
echo    ├── 启动.bat          ← 双击启动
echo    └── 首次使用-安装前端依赖.bat
echo.
echo ============================================================
echo.
echo  客户使用方式:
echo    1. 首次使用: 双击"首次使用-安装前端依赖.bat"
echo    2. 之后启动: 双击"启动.bat"
echo.
echo  是否压缩为 ZIP？
choice /C YN /M "压缩"
if errorlevel 2 goto :done

echo.
echo  正在压缩...
cd /d "%USERPROFILE%\Desktop"
powershell -Command "Compress-Archive -Path '%PACK_NAME%' -DestinationPath '%PACK_NAME%.zip' -Force"
echo.
echo  [√] ZIP 已创建: %USERPROFILE%\Desktop\%PACK_NAME%.zip
echo.

:done
echo ============================================================
pause

