#!/usr/bin/env python3
"""
OCR功能测试脚本
用于测试OCR服务器的文字识别功能
"""

import base64
import json
import requests
import sys
import os

def test_ocr_health():
    """测试OCR服务器健康状态"""
    try:
        response = requests.get('http://localhost:8080/health', timeout=5)
        print(f"✅ OCR服务器健康检查: {response.status_code}")
        if response.status_code == 200:
            print(f"   响应: {response.text}")
        return True
    except requests.exceptions.RequestException as e:
        print(f"❌ OCR服务器健康检查失败: {e}")
        return False

def test_ocr_api_info():
    """测试OCR API信息接口"""
    try:
        response = requests.get('http://localhost:8080/', timeout=5)
        print(f"✅ OCR API信息: {response.status_code}")
        if response.status_code == 200:
            print(f"   响应: {response.text}")
        return True
    except requests.exceptions.RequestException as e:
        print(f"❌ OCR API信息获取失败: {e}")
        return False

def test_ocr_recognition(image_path):
    """测试OCR文字识别功能"""
    if not os.path.exists(image_path):
        print(f"❌ 图片文件不存在: {image_path}")
        return False
    
    try:
        # 读取图片文件
        with open(image_path, 'rb') as f:
            image_data = base64.b64encode(f.read()).decode('utf-8')
        
        # 准备请求数据
        request_data = {
            'image': image_data
        }
        
        # 发送OCR请求
        response = requests.post(
            'http://localhost:8080/ocr',
            json=request_data,
            timeout=30
        )
        
        print(f"✅ OCR识别请求: {response.status_code}")
        
        if response.status_code == 200:
            result = response.json()
            print(f"   识别结果:")
            print(f"   成功: {result.get('success', False)}")
            
            if result.get('success'):
                data = result.get('data', {})
                text_count = data.get('text_count', 0)
                results = data.get('results', [])
                
                print(f"   检测到 {text_count} 个文本区域")
                
                for i, item in enumerate(results, 1):
                    text = item.get('text', '')
                    confidence = item.get('confidence', 0)
                    bbox = item.get('bbox', {})
                    
                    print(f"   区域 {i}:")
                    print(f"     文本: {text}")
                    print(f"     置信度: {confidence:.2f}")
                    print(f"     边界框: ({bbox.get('xmin', 0)}, {bbox.get('ymin', 0)}) - ({bbox.get('xmax', 0)}, {bbox.get('ymax', 0)})")
            else:
                print(f"   OCR识别失败")
            
            return True
        else:
            print(f"   OCR请求失败: {response.text}")
            return False
            
    except requests.exceptions.RequestException as e:
        print(f"❌ OCR识别请求失败: {e}")
        return False
    except Exception as e:
        print(f"❌ OCR识别处理失败: {e}")
        return False

def main():
    """主函数"""
    print("OCR功能测试")
    print("=" * 50)
    
    # 测试健康检查
    print("\n1. 测试OCR服务器健康状态...")
    health_ok = test_ocr_health()
    
    # 测试API信息
    print("\n2. 测试OCR API信息...")
    api_ok = test_ocr_api_info()
    
    # 测试OCR识别
    print("\n3. 测试OCR文字识别...")
    test_image = "test.png"
    if os.path.exists(test_image):
        recognition_ok = test_ocr_recognition(test_image)
    else:
        print(f"⚠️  测试图片不存在: {test_image}")
        print("   请准备一个测试图片文件")
        recognition_ok = False
    
    # 总结
    print("\n" + "=" * 50)
    print("测试总结:")
    print(f"   健康检查: {'✅ 通过' if health_ok else '❌ 失败'}")
    print(f"   API信息: {'✅ 通过' if api_ok else '❌ 失败'}")
    print(f"   文字识别: {'✅ 通过' if recognition_ok else '❌ 失败'}")
    
    if health_ok and api_ok and recognition_ok:
        print("\n🎉 所有测试通过！OCR功能正常工作。")
        return 0
    else:
        print("\n⚠️  部分测试失败，请检查OCR服务器状态。")
        return 1

if __name__ == "__main__":
    sys.exit(main()) 