#!/bin/bash
echo "OCR服务器启动中..."
echo "使用示例:"
echo "  ./start.sh                    # 使用默认端口8080"
echo "  ./start.sh -p 9000           # 使用端口9000"
echo "  ./start.sh --port 9000       # 使用端口9000"
echo "  ./start.sh -H 0.0.0.0 -p 9000  # 绑定到所有接口的9000端口"
echo "  ./start.sh --auto-port       # 自动查找可用端口"
echo "  ./start.sh --help            # 显示帮助信息"
echo ""
if [ $# -eq 0 ]; then
    ./ocr_server
else
    ./ocr_server "$@"
fi
