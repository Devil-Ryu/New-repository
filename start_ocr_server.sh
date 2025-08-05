#!/bin/bash

echo "启动OCR服务器..."
echo "=================="

# 检查OCR服务器文件是否存在
if [ ! -f "./ocr_minimal/ocr_server" ]; then
    echo "错误: OCR服务器文件不存在: ./ocr_minimal/ocr_server"
    echo "请确保OCR服务器已正确编译"
    exit 1
fi

# 检查OCR服务器是否可执行
if [ ! -x "./ocr_minimal/ocr_server" ]; then
    echo "错误: OCR服务器文件不可执行"
    echo "请运行: chmod +x ./ocr_minimal/ocr_server"
    exit 1
fi

# 启动OCR服务器
echo "正在启动OCR服务器..."
cd ocr_minimal
./start.sh

echo ""
echo "OCR服务器已启动，默认地址: http://localhost:8080"
echo "您可以使用以下命令测试OCR服务:"
echo "  curl -X POST http://localhost:8080/health"
echo ""
echo "按 Ctrl+C 停止服务器" 