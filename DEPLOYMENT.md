# 部署指南

本文档详细说明如何将考试小助手项目部署到GitHub并使用GitHub Actions进行自动化构建。

## 准备工作

### 1. 环境要求
- Git 2.0+
- Go 1.24+
- Node.js 18+
- GitHub账户

### 2. 本地开发环境设置
```bash
# 安装Wails CLI
go install github.com/wailsapp/wails/v3/cmd/wails@latest

# 验证安装
wails doctor
```

## GitHub仓库设置

### 1. 创建GitHub仓库
1. 登录GitHub
2. 点击右上角"+"号，选择"New repository"
3. 填写仓库信息：
   - Repository name: `exam_assistant`
   - Description: `考试小助手 - 基于Wails的跨平台桌面应用`
   - 选择Public或Private
   - **不要**勾选"Add a README file"
   - **不要**勾选"Add .gitignore"
   - **不要**勾选"Choose a license"

### 2. 初始化本地仓库
```bash
# 在项目根目录执行
git init
git add .
git commit -m "Initial commit: 考试小助手项目"

# 添加远程仓库
git remote add origin https://github.com/your-username/exam_assistant.git
git branch -M main
git push -u origin main
```

## GitHub Actions配置

### 1. 工作流文件说明

项目包含两个GitHub Actions工作流：

#### `build.yml` - 构建和发布
- **触发条件**: 推送标签（v*）或手动触发
- **功能**: 多平台构建和自动发布
- **支持平台**: 
  - macOS Intel/ARM64
  - Windows AMD64/ARM64
  - Linux AMD64/ARM64

#### `test.yml` - 测试和代码检查
- **触发条件**: 推送到main/develop分支或PR
- **功能**: 代码质量检查和测试

### 2. 启用Actions
1. 推送代码后，GitHub会自动检测到`.github/workflows/`目录
2. 在仓库页面点击"Actions"标签
3. 确认工作流已启用

## 构建和发布流程

### 1. 开发流程
```bash
# 创建功能分支
git checkout -b feature/new-feature

# 开发完成后提交
git add .
git commit -m "feat: 添加新功能"

# 推送到远程
git push origin feature/new-feature

# 创建Pull Request
# 在GitHub上创建PR，合并到main分支
```

### 2. 发布新版本
```bash
# 确保在main分支
git checkout main
git pull origin main

# 创建版本标签
git tag v1.0.0
git push origin v1.0.0
```

### 3. 监控构建过程
1. 推送标签后，GitHub Actions会自动触发构建
2. 在"Actions"页面查看构建进度
3. 构建完成后会自动创建Release

## 构建产物

### 1. 构建输出
- **macOS**: `.app`文件和`.dmg`安装包
- **Windows**: `.exe`文件和`.msi`安装包
- **Linux**: 可执行文件和`.AppImage`

### 2. 发布文件
构建完成后，以下文件会自动上传到GitHub Release：
- `exam_assistant-darwin_amd64/` - macOS Intel版本
- `exam_assistant-darwin_arm64/` - macOS Apple Silicon版本
- `exam_assistant-windows_amd64/` - Windows AMD64版本
- `exam_assistant-windows_arm64/` - Windows ARM64版本
- `exam_assistant-linux_amd64/` - Linux AMD64版本
- `exam_assistant-linux_arm64/` - Linux ARM64版本

## 故障排除

### 1. 常见问题

#### 构建失败
- 检查Go版本是否为1.24+
- 确认Node.js版本为18+
- 查看Actions日志获取详细错误信息

#### 依赖问题
```bash
# 清理Go模块缓存
go clean -modcache

# 重新下载依赖
go mod tidy
go mod download
```

#### 前端构建问题
```bash
# 清理前端缓存
cd frontend
rm -rf node_modules package-lock.json
npm install
npm run build
```

### 2. 调试技巧
- 在Actions页面查看详细日志
- 使用`workflow_dispatch`手动触发构建进行测试
- 检查`.gitignore`确保必要文件被包含

## 自定义配置

### 1. 修改应用信息
编辑`wails.json`文件：
```json
{
  "name": "你的应用名称",
  "author": {
    "name": "你的姓名",
    "email": "你的邮箱"
  }
}
```

### 2. 调整构建配置
在`.github/workflows/build.yml`中：
- 修改Go版本
- 调整Node.js版本
- 添加自定义构建步骤

### 3. 添加图标
1. 准备应用图标文件（建议512x512 PNG）
2. 放置到`build/appicon.png`
3. 重新构建应用

## 最佳实践

### 1. 版本管理
- 使用语义化版本号（如v1.0.0）
- 在README中维护更新日志
- 为每个版本添加详细的Release说明

### 2. 代码质量
- 提交前运行本地测试
- 使用有意义的提交信息
- 定期更新依赖包

### 3. 文档维护
- 及时更新README.md
- 维护API文档
- 记录重要的配置变更

## 联系支持

如果遇到问题，请：
1. 查看GitHub Issues
2. 提交详细的错误报告
3. 提供复现步骤和环境信息 