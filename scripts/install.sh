#!/bin/bash
echo "============================================================"
echo "    superAIAutoCutVideo 安装脚本 (macOS/Linux)"
echo "    使用国内镜像源 - 无需 VPN"
echo "============================================================"
echo

# 国内镜像配置
PIP_MIRROR="https://pypi.tuna.tsinghua.edu.cn/simple"
NPM_MIRROR="https://registry.npmmirror.com"

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# 检查 Python
echo "[环境检查]"
if ! command -v python3 &> /dev/null; then
    echo "  [X] 未检测到 Python3，请先安装"
    exit 1
fi
echo "  [√] Python3: $(python3 --version)"

# 检查 Node.js
if ! command -v node &> /dev/null; then
    echo "  [X] 未检测到 Node.js，请先安装"
    exit 1
fi
echo "  [√] Node.js: $(node --version)"

echo "  [√] 使用国内镜像源（无需 VPN）"
echo "      pip: $PIP_MIRROR"
echo "      npm: $NPM_MIRROR"
echo

echo "[1/6] 创建 Python 虚拟环境..."
cd "$PROJECT_DIR/backend"
if [ -d "venv" ]; then
    echo "     虚拟环境已存在，跳过创建"
else
    python3 -m venv venv
    echo "     [√] 虚拟环境创建成功"
fi

echo
echo "[2/6] 配置 pip 国内镜像并升级..."
source venv/bin/activate
pip config set global.index-url $PIP_MIRROR 2>/dev/null
pip install --upgrade pip -i $PIP_MIRROR -q
echo "     [√] pip 已配置清华镜像"

echo
echo "[3/6] 安装后端 Python 依赖..."
echo "     这可能需要几分钟，请耐心等待..."
pip install -r requirements.txt -i $PIP_MIRROR
echo "     [√] 后端依赖安装完成"

echo
echo "[4/6] 安装 IndexTTS2..."
echo "     正在从 Gitee 镜像安装（国内加速）..."
pip install git+https://gitee.com/mirrors/index-tts.git -i $PIP_MIRROR 2>/dev/null
if [ $? -ne 0 ]; then
    echo "     Gitee 镜像失败，尝试 GitHub..."
    pip install git+https://github.com/index-tts/index-tts.git -i $PIP_MIRROR 2>/dev/null
    if [ $? -ne 0 ]; then
        echo "     [!] IndexTTS2 安装失败，请稍后手动安装"
    else
        echo "     [√] IndexTTS2 安装完成（GitHub）"
    fi
else
    echo "     [√] IndexTTS2 安装完成（Gitee 镜像）"
fi

echo
echo "[5/6] 安装前端 npm 依赖..."
cd "$PROJECT_DIR/frontend"
echo "     配置 npm 淘宝镜像..."
npm config set registry $NPM_MIRROR
echo "     这可能需要几分钟，请耐心等待..."
npm install
echo "     [√] 前端依赖安装完成"

echo
echo "[6/6] 下载 IndexTTS2 模型..."
cd "$PROJECT_DIR/backend"
source venv/bin/activate

# 检查模型是否已存在
if [ -f "checkpoints/config.yaml" ]; then
    echo "     模型已存在，跳过下载"
else
    echo "     模型大小约 3-5GB，请确保网络畅通..."
    echo "     正在从 ModelScope 下载（国内服务器，无需 VPN）..."
    
    # 先安装 modelscope
    pip install modelscope -i $PIP_MIRROR -q
    
    python3 -c "from modelscope import snapshot_download; snapshot_download('IndexTeam/IndexTTS-1.5', local_dir='./checkpoints')"
    if [ $? -ne 0 ]; then
        echo "     [!] 模型下载失败"
        echo "     请手动下载: https://modelscope.cn/models/IndexTeam/IndexTTS-1.5"
    else
        echo "     [√] 模型下载完成"
    fi
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
echo "使用的国内镜像:"
echo "  pip:  清华大学 ($PIP_MIRROR)"
echo "  npm:  淘宝镜像 ($NPM_MIRROR)"
echo "  模型: ModelScope（阿里云）"
echo
echo "注意: 首次使用需准备音色文件，详见使用指南"
echo "============================================================"
