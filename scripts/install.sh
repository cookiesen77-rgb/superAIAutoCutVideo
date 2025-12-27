#!/bin/bash
echo "============================================================"
echo "    superAIAutoCutVideo 安装脚本 (macOS/Linux)"
echo "============================================================"
echo

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# 检查 Python
if ! command -v python3 &> /dev/null; then
    echo "[错误] 未检测到 Python3，请先安装"
    exit 1
fi

# 检查 Node.js
if ! command -v node &> /dev/null; then
    echo "[错误] 未检测到 Node.js，请先安装"
    exit 1
fi

echo "[1/5] 创建 Python 虚拟环境..."
cd "$PROJECT_DIR/backend"
if [ -d "venv" ]; then
    echo "     虚拟环境已存在，跳过创建"
else
    python3 -m venv venv
    echo "     虚拟环境创建成功"
fi

echo
echo "[2/5] 激活虚拟环境并安装后端依赖..."
source venv/bin/activate
pip install --upgrade pip -q
pip install -r requirements.txt -q
echo "     后端依赖安装完成"

echo
echo "[3/5] 安装 IndexTTS2..."
pip install git+https://github.com/index-tts/index-tts.git -q 2>/dev/null
if [ $? -ne 0 ]; then
    echo "     [警告] IndexTTS2 安装失败，请稍后手动安装"
else
    echo "     IndexTTS2 安装完成"
fi

echo
echo "[4/5] 安装前端依赖..."
cd "$PROJECT_DIR/frontend"
npm install --silent
echo "     前端依赖安装完成"

echo
echo "[5/5] 下载 IndexTTS2 模型..."
cd "$PROJECT_DIR/backend"
source venv/bin/activate
python3 -c "from modelscope import snapshot_download; snapshot_download('IndexTeam/IndexTTS-1.5', local_dir='./checkpoints')" 2>/dev/null
if [ $? -ne 0 ]; then
    echo "     [警告] 模型下载失败，请稍后手动下载"
    echo "     参考文档: docs/IndexTTS2_使用指南.md"
else
    echo "     模型下载完成"
fi

echo
echo "============================================================"
echo "    安装完成！"
echo "============================================================"
echo
echo "启动方式:"
echo "  ./scripts/start.sh"
echo
echo "或手动启动:"
echo "  # 终端1 - 后端"
echo "  cd backend && source venv/bin/activate && python main.py"
echo
echo "  # 终端2 - 前端"  
echo "  cd frontend && npm run dev"
echo
echo "注意: 首次使用需准备音色文件，详见使用指南"
echo "============================================================"

