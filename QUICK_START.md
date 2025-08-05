# 快速上手指南

## 🚀 5分钟部署到GitHub

### 第一步：准备项目
```bash
# 确保在项目根目录
cd /Users/chuva/Projects/Wails_Projects/exam_assistant

# 初始化Git仓库
git init
git add .
git commit -m "Initial commit: 考试小助手项目"
```

### 第二步：创建GitHub仓库
1. 访问 [GitHub](https://github.com)
2. 点击右上角"+" → "New repository"
3. 填写信息：
   - Repository name: `exam_assistant`
   - Description: `考试小助手 - 基于Wails的跨平台桌面应用`
   - 选择Public
   - **不要**勾选任何初始化选项
4. 点击"Create repository"

### 第三步：推送代码
```bash
# 替换your-username为你的GitHub用户名
git remote add origin https://github.com/your-username/exam_assistant.git
git branch -M main
git push -u origin main
```

### 第四步：触发首次构建
```bash
# 创建第一个版本标签
git tag v1.0.0
git push origin v1.0.0
```

### 第五步：查看构建结果
1. 访问你的GitHub仓库页面
2. 点击"Actions"标签
3. 查看构建进度
4. 构建完成后会自动创建Release

## 📋 检查清单

### 环境检查
- [ ] Go 1.24+ 已安装
- [ ] Node.js 18+ 已安装
- [ ] Git 已配置
- [ ] GitHub账户已创建

### 文件检查
- [ ] `.github/workflows/build.yml` 存在
- [ ] `.github/workflows/test.yml` 存在
- [ ] `wails.json` 已配置
- [ ] `.gitignore` 已设置
- [ ] `README.md` 已更新

### 构建检查
- [ ] 本地测试通过：`wails dev`
- [ ] 前端构建成功：`cd frontend && npm run build`
- [ ] GitHub Actions已启用
- [ ] 首次构建成功

## 🔧 常见问题解决

### 问题1：构建失败
```bash
# 检查Go版本
go version

# 检查Node.js版本
node --version

# 清理并重新安装依赖
cd frontend
rm -rf node_modules package-lock.json
npm install
```

### 问题2：GitHub Actions未触发
1. 检查仓库设置 → Actions → General
2. 确保"Allow all actions and reusable workflows"已启用
3. 检查工作流文件语法是否正确

### 问题3：构建产物缺失
1. 检查`.gitignore`是否排除了必要文件
2. 确认`wails.json`配置正确
3. 查看Actions日志获取详细错误信息

## 📱 支持的平台

| 平台 | 架构 | 状态 |
|------|------|------|
| macOS | Intel (x64) | ✅ |
| macOS | Apple Silicon (ARM64) | ✅ |
| Windows | AMD64 | ✅ |
| Windows | ARM64 | ✅ |
| Linux | AMD64 | ✅ |
| Linux | ARM64 | ✅ |

## 🎯 下一步

1. **自定义应用信息**：编辑`wails.json`中的应用名称和作者信息
2. **添加应用图标**：将图标文件放置到`build/appicon.png`
3. **配置自动更新**：考虑添加自动更新机制
4. **添加测试**：编写单元测试和集成测试
5. **文档完善**：添加用户手册和API文档

## 📞 获取帮助

- 📖 查看详细文档：[DEPLOYMENT.md](./DEPLOYMENT.md)
- 🐛 报告问题：[GitHub Issues](https://github.com/your-username/exam_assistant/issues)
- 💬 讨论功能：[GitHub Discussions](https://github.com/your-username/exam_assistant/discussions)

---

**恭喜！** 🎉 你的项目现在已经配置了完整的CI/CD流程，支持多平台自动化构建和发布。 