# OCR功能使用说明

## 概述

本项目已集成OCR文字识别功能，使用本地OCR服务器进行文字识别。OCR服务器基于ONNX模型，支持中文文字识别。

## 启动步骤

### 1. 启动OCR服务器

在项目根目录运行：

```bash
./start_ocr_server.sh
```

或者手动启动：

```bash
cd ocr_minimal
./start.sh
```

OCR服务器将在 `http://localhost:8080` 启动。

### 2. 启动主应用程序

```bash
go run .
```

或者编译后运行：

```bash
go build -o exam_assistant .
./exam_assistant
```

## OCR功能特性

### 支持的接口

1. **POST /ocr** - 文字识别接口
   - 接收base64编码的图片
   - 返回JSON格式的识别结果

2. **GET /health** - 健康检查接口
   - 检查OCR服务状态

3. **GET /** - API信息接口
   - 获取API使用说明

### 识别结果格式

```json
{
    "success": true,
    "data": {
        "text_count": 2,
        "results": [
            {
                "text": "识别的文字",
                "confidence": 0.95,
                "bbox": {
                    "xmin": 100,
                    "ymin": 50,
                    "xmax": 300,
                    "ymax": 80,
                    "points": [[100, 50], [300, 50], [300, 80], [100, 80]]
                }
            }
        ]
    }
}
```

## 测试OCR功能

### 1. 测试OCR服务器

```bash
# 健康检查
curl http://localhost:8080/health

# 获取API信息
curl http://localhost:8080/
```

### 2. 测试OCR识别

可以使用提供的测试图片：

```bash
# 将图片转换为base64并发送到OCR服务
python3 -c "
import base64
import requests
import json

# 读取图片文件
with open('test.png', 'rb') as f:
    image_data = base64.b64encode(f.read()).decode('utf-8')

# 发送OCR请求
response = requests.post('http://localhost:8080/ocr', 
                        json={'image': image_data})
print(json.dumps(response.json(), indent=2, ensure_ascii=False))
"
```

## 故障排除

### 1. OCR服务器启动失败

- 检查 `./ocr_minimal/ocr_server` 文件是否存在
- 确保文件有执行权限：`chmod +x ./ocr_minimal/ocr_server`
- 检查端口8080是否被占用

### 2. 主应用程序无法连接OCR服务

- 确保OCR服务器正在运行
- 检查防火墙设置
- 确认服务器地址为 `http://localhost:8080`

### 3. OCR识别结果不准确

- 确保图片清晰度足够
- 检查图片格式是否支持（推荐PNG格式）
- 尝试调整图片大小或对比度

## 配置选项

### OCR服务器配置

可以通过修改 `ocr_minimal/start.sh` 来调整服务器配置：

```bash
# 使用不同端口
./start.sh -p 9000

# 绑定到所有接口
./start.sh -H 0.0.0.0 -p 8080

# 自动查找可用端口
./start.sh --auto-port
```

### 主应用程序配置

在 `main.go` 中可以修改OCR服务器地址：

```go
ocrServerURL := "http://localhost:8080"  // 修改为实际地址
```

## 性能优化

1. **图片预处理**：在发送到OCR服务前对图片进行预处理
2. **缓存结果**：对相同图片的识别结果进行缓存
3. **批量处理**：支持批量图片识别
4. **异步处理**：使用goroutine进行异步OCR处理

## 扩展功能

可以考虑添加的功能：

- 多语言支持
- 手写文字识别
- 表格识别
- 文档OCR
- 实时OCR预览 