#!/bin/bash
echo "============================================================"
echo "    superAIAutoCutVideo 启动脚本"
echo "============================================================"
echo

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# 检查虚拟环境
if [ ! -f "$PROJECT_DIR/backend/venv/bin/activate" ]; then
    echo "[错误] 虚拟环境不存在，请先运行 ./scripts/install.sh"
    exit 1
fi

# 启动后端
echo "[启动后端服务...]"
cd "$PROJECT_DIR/backend"
source venv/bin/activate
python main.py &
BACKEND_PID=$!
echo "     后端 PID: $BACKEND_PID"

# 等待后端启动
echo "等待后端启动 (5秒)..."
sleep 5

# 启动前端
echo "[启动前端服务...]"
cd "$PROJECT_DIR/frontend"
npm run dev &
FRONTEND_PID=$!
echo "     前端 PID: $FRONTEND_PID"

echo
echo "============================================================"
echo "    服务已启动！"
echo "============================================================"
echo
echo "后端地址: http://localhost:8000"
echo "前端地址: http://localhost:5173"
echo
echo "按 Ctrl+C 停止所有服务"
echo "============================================================"

# 捕获退出信号
trap "echo '正在停止服务...'; kill $BACKEND_PID $FRONTEND_PID 2>/dev/null; exit" SIGINT SIGTERM

# 保持脚本运行
wait

