<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ExamService } from '../bindings/changeme/index.js'

// 响应式数据
const leftPanelWidth = ref(600) // 默认占50% (1200px的一半)
const ocrConfig = reactive({
  mode: 'local',
  url: 'http://127.0.0.1:8080',
  apiKey: '',
  status: '未链接'
})

const importConfig = reactive({
  fileType: 'csv',
  encoding: 'utf8',
  answerDelimiter: '\\n',
  optionDelimiter: '\\n'
})

const screenshotArea = reactive({
  x: 0,
  y: 0,
  width: 0,
  height: 0,
  image: ''
})

const selectedAreaImage = ref('')
const ocrResult = ref('')
const answers = ref([])
const searchResults = ref([])
const selectedTypeFilters = ref([]) // 选中的题目类型筛选（多选）
const selectedSearchTypeFilters = ref([]) // 搜索结果中选中的题目类型筛选（多选）

// 计算属性：筛选后的答案列表
const filteredAnswers = computed(() => {
  // 如果没有选择任何类型，显示所有答案
  if (selectedTypeFilters.value.length === 0) {
    return answers.value
  }
  // 否则只显示选中类型的答案
  return answers.value.filter(answer => selectedTypeFilters.value.includes(answer.type))
})

// 计算属性：可用的题目类型
const availableTypes = computed(() => {
  const types = new Set(answers.value.map(answer => answer.type))
  return Array.from(types).sort()
})

// 计算属性：搜索结果中可用的题目类型
const availableSearchTypes = computed(() => {
  const types = new Set(searchResults.value.map(result => result.item.type))
  return Array.from(types).sort()
})

// 计算属性：筛选后的搜索结果
const filteredSearchResults = computed(() => {
  // 如果没有选择任何类型，显示所有搜索结果
  if (selectedSearchTypeFilters.value.length === 0) {
    return searchResults.value
  }
  // 否则只显示选中类型的搜索结果
  return searchResults.value.filter(result => selectedSearchTypeFilters.value.includes(result.item.type))
})

// 处理筛选器变化
const handleFilterChange = (value) => {
  selectedTypeFilters.value = value
}

// 处理搜索结果筛选器变化
const handleSearchFilterChange = (value) => {
  selectedSearchTypeFilters.value = value
}

// 错误弹窗相关
const showErrorDialog = ref(false)
const errorDialogTitle = ref('')
const errorDialogContent = ref('')
const errorDialogDetails = ref('')

// 显示错误弹窗
const showError = (title, content, details = '') => {
  errorDialogTitle.value = title
  errorDialogContent.value = content
  errorDialogDetails.value = details
  showErrorDialog.value = true
}

// 区域选择相关
const showAreaSelector = ref(false)
const fullScreenshot = ref('')
const selectedArea = ref(null)
const isSelecting = ref(false)
const startPoint = ref({ x: 0, y: 0 })
const imageRef = ref(null)

// 获取图片缩放比例
const getImageScale = () => {
  if (!imageRef.value) return { x: 1, y: 1 }
  
  const img = imageRef.value
  if (img.naturalWidth && img.offsetWidth) {
    return {
      x: img.naturalWidth / img.offsetWidth,
      y: img.naturalHeight / img.offsetHeight
    }
  }
  return { x: 1, y: 1 }
}

// 开始选择区域
const startSelection = (event) => {
  if (!showAreaSelector.value) return
  
  // 阻止默认行为
  event.preventDefault()
  event.stopPropagation()
  
  const rect = event.target.getBoundingClientRect()
  
  // 使用响应式的缩放比例
  const x = (event.clientX - rect.left) * getImageScale().x
  const y = (event.clientY - rect.top) * getImageScale().y
  
  startPoint.value = { 
    x: Math.round(x), 
    y: Math.round(y) 
  }
  isSelecting.value = true
  
  selectedArea.value = {
    x: Math.round(x),
    y: Math.round(y),
    width: 0,
    height: 0
  }
}

// 更新选择区域
const updateSelection = (event) => {
  if (!isSelecting.value || !selectedArea.value) return
  
  // 阻止默认行为
  event.preventDefault()
  event.stopPropagation()
  
  const rect = event.target.getBoundingClientRect()
  
  // 使用响应式的缩放比例
  const x = (event.clientX - rect.left) * getImageScale().x
  const y = (event.clientY - rect.top) * getImageScale().y
  
  const startX = Math.min(startPoint.value.x, x)
  const startY = Math.min(startPoint.value.y, y)
  const width = Math.abs(x - startPoint.value.x)
  const height = Math.abs(y - startPoint.value.y)
  
  // 确保坐标值为整数，避免精度问题
  selectedArea.value = {
    x: Math.round(startX),
    y: Math.round(startY),
    width: Math.round(width),
    height: Math.round(height)
  }
}

// 结束选择区域
const endSelection = (event) => {
  if (event) {
    event.preventDefault()
    event.stopPropagation()
  }
  isSelecting.value = false
}

// 测试本地OCR连接
const testLocalOCRConnection = async () => {
  try {
    console.log('开始测试本地OCR连接')
    ocrConfig.status = '连接中'
    
    // 通过后端代理测试OCR连接
    const { ExamService } = await import('../bindings/changeme/index.js')
    const result = await ExamService.TestOCRConnection(ocrConfig)
    
    if (result === '连接成功') {
      ocrConfig.status = '连接成功'
      console.log('OCR服务连接测试成功')
    } else {
      ocrConfig.status = '连接失败'
      console.error('OCR服务连接测试失败:', result)
    }
  } catch (error) {
    console.error('本地OCR连接测试失败:', error)
    ocrConfig.status = '连接失败'
  }
}

// 测试本地OCR功能（已移除，使用testLocalOCRConnection替代）

// 导入答案
const importAnswers = async () => {
  console.log('导入答案按钮被点击')
  
  try {
    // 使用Wails绑定文件
    const { ExamService } = await import('../bindings/changeme/index.js')
    
    // 根据文件类型设置对话框标题和文件类型
    const fileType = importConfig.fileType === 'csv' ? 'csv' : 'excel'
    const dialogTitle = importConfig.fileType === 'csv' ? '选择CSV答案文件' : '选择Excel答案文件'
    
    // 打开文件对话框
    const result = await ExamService.OpenFileDialog(dialogTitle, fileType)
    
    if (result.success && result.filePath) {
      console.log('选择的文件路径:', result.filePath)
      
      try {
        // 根据文件类型调用不同的导入方法
        console.log(`开始导入${importConfig.fileType.toUpperCase()}文件:`, result.filePath)
        let newAnswers
        
        if (importConfig.fileType === 'csv') {
          newAnswers = await ExamService.ParseCSVFile(result.filePath, importConfig.encoding, importConfig.optionDelimiter, importConfig.answerDelimiter)
        } else {
          // TODO: 添加Excel文件导入支持
          throw new Error('Excel文件导入功能暂未实现')
        }
        
        // 验证解析结果
        if (!newAnswers || newAnswers.length === 0) {
          throw new Error('文件中没有找到有效的答案数据')
        }
        
        // 更新答案数据
        answers.value = newAnswers
        console.log('导入成功，共导入', newAnswers.length, '条答案')
        
        // 清空搜索结果
        searchResults.value = []
        
        // 显示成功提示
        alert(`成功导入 ${newAnswers.length} 条答案数据！`)
      } catch (error) {
        console.error('文件导入失败:', error)
        
        // 根据错误类型显示详细的错误信息
        let errorTitle = '文件导入失败'
        let errorContent = ''
        let errorDetails = ''
        
        if (error.message.includes('无法打开CSV文件')) {
          errorTitle = 'CSV文件访问失败'
          errorContent = '无法打开CSV文件，请检查文件是否存在且可访问'
          errorDetails = '可能的原因：\n• 文件路径不正确\n• 文件权限不足\n• 文件被其他程序占用'
        } else if (error.message.includes('编码设置错误')) {
          errorTitle = '文件编码错误'
          errorContent = '文件编码设置错误，请尝试其他编码格式'
          errorDetails = '支持的编码格式：\n• UTF-8（推荐）\n• GBK（中文Windows）'
        } else if (error.message.includes('读取CSV文件失败')) {
          errorTitle = 'CSV文件读取失败'
          errorContent = '读取CSV文件失败，请检查文件是否损坏'
          errorDetails = '可能的原因：\n• 文件已损坏\n• 磁盘空间不足\n• 文件权限问题'
        } else if (error.message.includes('文件缺少必需字段')) {
          errorTitle = '文件格式错误'
          errorContent = `文件缺少必需字段：${error.message.replace('文件缺少必需字段: ', '')}`
          errorDetails = `请确保CSV文件第一行包含以下列名：\n• 类型\n• 题目\n• 选项\n• 答案\n\n注意：列名必须包含这些关键词，顺序可以不同`
        } else if (error.message.includes('没有找到有效的数据行')) {
          errorTitle = '数据行无效'
          errorContent = '文件中没有找到有效的数据行'
          errorDetails = '请检查：\n• 文件是否包含数据行（第二行开始）\n• 数据格式是否正确\n• 必需字段（类型、题目）是否填写\n• 数据行是否为空'
        } else if (error.message.includes('文件为空')) {
          errorTitle = '文件内容为空'
          errorContent = '选择的文件为空，请选择包含数据的文件'
          errorDetails = '请选择包含答案数据的文件'
        } else if (error.message.includes('文件缺少标题行')) {
          errorTitle = '文件缺少标题行'
          errorContent = '文件缺少标题行，请确保CSV文件第一行包含列名'
          errorDetails = '请确保CSV文件第一行包含以下列名：\n• 类型\n• 题目\n• 选项\n• 答案\n\n注意：列名必须包含这些关键词，顺序可以不同'
        } else if (error.message.includes('解析CSV文件失败')) {
          errorTitle = 'CSV文件解析失败'
          errorContent = 'CSV文件格式错误，请检查文件内容'
          errorDetails = '请检查：\n• CSV文件格式是否正确\n• 列分隔符是否正确\n• 数据是否完整'
        } else {
          errorTitle = '文件导入失败'
          errorContent = '文件导入过程中发生未知错误'
          errorDetails = `错误详情：${error.message}`
        }
        
        showError(errorTitle, errorContent, errorDetails)
      }
    } else {
      console.log('用户取消了文件选择')
    }
  } catch (error) {
    console.error('导入失败:', error)
    showError('导入失败', '导入过程中发生错误', `错误详情：${error.message}`)
  }
}

// 读取文件为文本
const readFileAsText = (file) => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    
    reader.onload = (e) => {
      resolve(e.target.result)
    }
    
    reader.onerror = () => {
      reject(new Error('文件读取失败'))
    }
    
    // 根据编码读取文件
    if (importConfig.encoding === 'utf8') {
      reader.readAsText(file, 'UTF-8')
    } else if (importConfig.encoding === 'gbk') {
      reader.readAsText(file, 'GBK')
    } else {
      reader.readAsText(file, 'UTF-8')
    }
  })
}

// 解析文件内容
const parseFileContent = async (content, filePath) => {
  try {
    let answers = []
    
    if (importConfig.fileType === 'csv') {
      // 解析CSV
      const lines = content.split('\n')
      console.log('CSV文件行数:', lines.length)
      
      for (let i = 1; i < lines.length; i++) {
        if (lines[i].trim()) {
          const values = lines[i].split(',').map(v => v.trim())
          console.log('解析行:', values)
          
          const answer = {
            type: values[0] || '未知类型',
            question: values[1] || '',
            options: values[2] ? values[2].split(',').map(opt => opt.trim()) : [],
            answer: values[3] ? values[3].split(',').map(ans => ans.trim()) : []
          }
          answers.push(answer)
        }
      }
    } else {
      // 解析Excel (模拟)
      answers = [
        {
          type: '选择题',
          question: '什么是Vue.js？',
          options: ['A. 一个JavaScript框架', 'B. 一个数据库', 'C. 一个操作系统'],
          answer: ['A']
        },
        {
          type: '填空题',
          question: 'Vue.js的创始人是谁？',
          options: [],
          answer: ['尤雨溪']
        }
      ]
    }
    
    console.log('解析完成，共', answers.length, '条答案')
    return answers
  } catch (error) {
    console.error('文件内容解析失败:', error)
    throw error
  }
}

// 选择区域
const selectArea = async () => {
  try {
    // 使用Wails绑定文件
    const { ExamService } = await import('../bindings/changeme/index.js')
    
    // 使用带窗口控制的截图方法
    const screenshot = await ExamService.TakeScreenshotWithWindowControl()
    fullScreenshot.value = screenshot
    
    // 等待一小段时间确保窗口完全显示
    await new Promise(resolve => setTimeout(resolve, 300))
    
    // 显示区域选择器
    showAreaSelector.value = true
    
    // 初始化选择区域
    selectedArea.value = {
      x: 100,
      y: 100,
      width: 400,
      height: 200
    }
  } catch (error) {
    console.error('截图失败:', error)
    alert('截图失败: ' + error.message)
  }
}

// 隐藏应用窗口
const hideApplicationWindow = async () => {
  try {
    // 使用后端方法隐藏窗口
    const { ExamService } = await import('../bindings/changeme/index.js')
    await ExamService.HideWindow()
  } catch (error) {
    console.warn('无法隐藏窗口:', error)
    // 如果无法隐藏窗口，继续执行截图
  }
}

// 显示应用窗口
const showApplicationWindow = async () => {
  try {
    // 使用后端方法显示窗口
    const { ExamService } = await import('../bindings/changeme/index.js')
    await ExamService.ShowWindow()
  } catch (error) {
    console.warn('无法显示窗口:', error)
  }
}

// 裁剪图片用于主页显示
const cropImageForDisplay = (imageSrc, area) => {
  return new Promise((resolve, reject) => {
    const img = new Image()
    img.onload = () => {
      try {
        // 创建canvas
        const canvas = document.createElement('canvas')
        const ctx = canvas.getContext('2d')
        
        // 计算保持宽高比的显示尺寸
        const maxDisplayWidth = 300  // 最大显示宽度
        const maxDisplayHeight = 200 // 最大显示高度
        
        // 计算原始区域的宽高比
        const originalRatio = area.width / area.height
        const displayRatio = maxDisplayWidth / maxDisplayHeight
        
        let displayWidth, displayHeight
        
        if (originalRatio > displayRatio) {
          // 原始区域更宽，以宽度为准
          displayWidth = maxDisplayWidth
          displayHeight = maxDisplayWidth / originalRatio
        } else {
          // 原始区域更高，以高度为准
          displayHeight = maxDisplayHeight
          displayWidth = maxDisplayHeight * originalRatio
        }
        
        canvas.width = displayWidth
        canvas.height = displayHeight
        
        // 绘制裁剪的区域并缩放到显示尺寸
        ctx.drawImage(
          img,
          area.x, area.y, area.width, area.height,  // 源图片裁剪区域
          0, 0, displayWidth, displayHeight          // 目标canvas区域（保持宽高比）
        )
        
        // 转换为base64
        const croppedImageSrc = canvas.toDataURL('image/png')
        resolve(croppedImageSrc)
      } catch (error) {
        reject(error)
      }
    }
    img.onerror = () => reject(new Error('图片加载失败'))
    img.src = imageSrc
  })
}

// 取消区域选择
const cancelAreaSelection = () => {
  showAreaSelector.value = false
  selectedArea.value = null
}

// 确认区域选择
const confirmAreaSelection = async () => {
  if (selectedArea.value) {
    try {
      // 保存原始截图的区域信息
      Object.assign(screenshotArea, selectedArea.value)
      screenshotArea.image = fullScreenshot.value  // 保存完整的原始截图
      
      // 生成适合主页显示的裁剪图片
      selectedAreaImage.value = await cropImageForDisplay(fullScreenshot.value, selectedArea.value)
      
      showAreaSelector.value = false
      
      console.log('区域选择完成:', selectedArea.value)
    } catch (error) {
      console.error('区域选择失败:', error)
      alert('区域选择失败: ' + error.message)
    }
  }
}

// 测试函数
const testFunction = () => {
  console.log('测试函数被调用')
  alert('测试函数被调用')
}

// 搜索答案
const searchAnswers = async () => {
  console.log('搜索按钮被点击')
  console.log('OCR结果:', ocrResult.value)
  console.log('答案数量:', answers.value.length)
  
  try {
    // 如果搜索内容为空，显示所有答案
    if (!ocrResult.value.trim()) {
      console.log('搜索内容为空，显示所有答案')
      searchResults.value = []
      return
    }
    
    console.log('开始搜索:', ocrResult.value)
    
    // 调用后端搜索方法
    const results = await ExamService.SearchAnswers(answers.value, ocrResult.value)
    console.log('后端返回结果:', results)
    
    // 显示所有匹配结果，按匹配度排序
    searchResults.value = results.sort((a, b) => b.score - a.score)
    
    console.log('搜索完成，找到', results.length, '条结果')
    
    if (results.length === 0) {
      alert('未找到匹配的答案')
    } else {
      // 显示匹配结果统计
      const highMatch = results.filter(r => r.score >= 0.8).length
      const mediumMatch = results.filter(r => r.score >= 0.5 && r.score < 0.8).length
      const lowMatch = results.filter(r => r.score < 0.5).length
      
      console.log(`匹配结果统计: 高匹配(${highMatch}个), 中匹配(${mediumMatch}个), 低匹配(${lowMatch}个)`)
    }
  } catch (error) {
    console.error('搜索失败:', error)
    alert('搜索失败: ' + error.message)
  }
}

// 下一题
const nextQuestion = async () => {
  try {
    console.log('开始下一题')
    
    // 检查是否已选择截图区域
    if (!screenshotArea.image) {
      alert('请先选择截图区域')
      return
    }
    
    // 1. 重新截取整个屏幕
    console.log('重新截取屏幕')
    const { ExamService } = await import('../bindings/changeme/index.js')
    const newScreenshot = await ExamService.TakeScreenshotWithWindowControl()
    
    // 2. 从新截图中提取选择区域
    console.log('从新截图中提取选择区域')
    const newAreaImage = await cropImageForDisplay(newScreenshot, screenshotArea)
    
    // 3. 更新主页上的截图
    selectedAreaImage.value = newAreaImage
    console.log('已更新主页截图')
    
    // 4. 进行OCR识别
    console.log('开始OCR识别')
    const ocrText = await performOCRWithBackend(newScreenshot, screenshotArea)
    ocrResult.value = ocrText
    console.log('OCR识别结果:', ocrText)
    
    // 5. 自动进行搜索
    console.log('开始自动搜索')
    await searchAnswers()
    
    console.log('下一题完成')
  } catch (error) {
    console.error('下一题失败:', error)
    alert('下一题失败: ' + error.message)
  }
}

// 通过后端执行OCR识别
const performOCRWithBackend = async (screenshotData, area) => {
  try {
    console.log('开始通过后端执行OCR识别')
    
    // 调用后端OCR识别功能
    const { ExamService } = await import('../bindings/changeme/index.js')
    
    // 创建截图区域对象
    const screenshotArea = {
      x: area.x,
      y: area.y,
      width: area.width,
      height: area.height,
      image: screenshotData
    }
    
    // 调用后端OCR识别
    const result = await ExamService.PerformOCR(screenshotArea, ocrConfig)
    
    if (result && result.trim()) {
      console.log('OCR识别成功:', result)
      return result
    } else {
      console.log('OCR识别返回空结果')
      return "OCR识别未返回有效内容"
    }
  } catch (error) {

    console.error('OCR识别失败:', error)
    return "OCR识别失败: " + error.message
  }
}

// 拖动调整宽度相关
const isResizing = ref(false)
const startX = ref(0)
const startWidth = ref(0)

// 开始拖动
const startResize = (e) => {
  isResizing.value = true
  startX.value = e.type === 'mousedown' ? e.clientX : e.touches[0].clientX
  startWidth.value = leftPanelWidth.value
  
  document.addEventListener('mousemove', handleResize)
  document.addEventListener('mouseup', stopResize)
  document.addEventListener('touchmove', handleResize)
  document.addEventListener('touchend', stopResize)
}

// 处理拖动
const handleResize = (e) => {
  if (!isResizing.value) return
  
  const currentX = e.type === 'mousemove' ? e.clientX : e.touches[0].clientX
  const deltaX = currentX - startX.value
  const newWidth = startWidth.value + deltaX
  
  // 限制最小和最大宽度 (30% - 70%)
  if (newWidth >= 360 && newWidth <= 840) {
    leftPanelWidth.value = newWidth
  }
}

// 停止拖动
const stopResize = () => {
  isResizing.value = false
  document.removeEventListener('mousemove', handleResize)
  document.removeEventListener('mouseup', stopResize)
  document.removeEventListener('touchmove', handleResize)
  document.removeEventListener('touchend', stopResize)
}

// 获取匹配度颜色
const getMatchScoreColor = (score) => {
  if (score >= 0.8) {
    return { background: '#f6ffed', color: '#52c41a', border: '#b7eb8f' } // 绿色
  } else if (score >= 0.5) {
    return { background: '#fff7e6', color: '#fa8c16', border: '#ffd591' } // 橙色
  } else {
    return { background: '#fff2f0', color: '#ff4d4f', border: '#ffccc7' } // 红色
  }
}

// 获取题目类型统计
const getTypeStats = () => {
  const stats = {}
  console.log('getTypeStats called, answers.length:', answers.value.length)
  answers.value.forEach(answer => {
    const type = answer.type
    stats[type] = (stats[type] || 0) + 1
  })
  console.log('Type stats:', stats)
  return stats
}

// 获取筛选后的题目类型统计
const getFilteredTypeStats = () => {
  const stats = {}
  filteredAnswers.value.forEach(answer => {
    const type = answer.type
    stats[type] = (stats[type] || 0) + 1
  })
  return stats
}

// 获取连接状态样式类
const getConnectionStatusClass = () => {
  if (ocrConfig.status === '连接成功') {
    return 'status-success'
  } else if (ocrConfig.status === '连接失败') {
    return 'status-error'
  } else if (ocrConfig.status === '连接中') {
    return 'status-connecting'
  } else {
    return 'status-default'
  }
}

// 高亮匹配的文本
const highlightText = (text, matches) => {
  // 直接返回原文本，不进行高亮处理
  return text
}

// 获取选择的区域图片
const getSelectedAreaImage = () => {
  if (!screenshotArea.image || !screenshotArea.width || !screenshotArea.height) {
    return null
  }
  
  return new Promise((resolve) => {
    const canvas = document.createElement('canvas')
    const ctx = canvas.getContext('2d')
    const img = new Image()
    
    img.onload = () => {
      canvas.width = screenshotArea.width
      canvas.height = screenshotArea.height
      
      ctx.drawImage(
        img,
        screenshotArea.x, screenshotArea.y, screenshotArea.width, screenshotArea.height,
        0, 0, screenshotArea.width, screenshotArea.height
      )
      
      resolve(canvas.toDataURL('image/png'))
    }
    img.src = screenshotArea.image
  })
}

// 初始化 - 答案页面默认为空
onMounted(() => {
  // 答案页面默认为空，用户需要导入数据
  answers.value = []
})
</script>

<template>
  <div class="app-container">
    <div class="main-layout">
      <!-- 左侧配置区域 -->
      <div class="config-panel" :style="{ width: leftPanelWidth + 'px' }">
        <div class="config-content">
          <!-- OCR配置区域 -->
          <div class="config-section">
            <h3>OCR配置</h3>
            <div class="config-row">
              <div class="config-item">
                <label class="config-label">OCR服务基础URL</label>
                <t-input
                  v-model="ocrConfig.url"
                  placeholder="请输入OCR服务基础URL，如: http://127.0.0.1:8080"
                  class="config-input"
                />
              </div>
            </div>
            <div class="config-row">
              <t-button @click="testLocalOCRConnection" theme="primary" variant="base" class="config-button">
                测试连接
              </t-button>
              <div class="connection-status-right">
                <t-tag 
                  :class="getConnectionStatusClass()"
                  class="status-tag"
                >
                  {{ ocrConfig.status || '未连接' }}
                </t-tag>
              </div>
            </div>
          </div>

          <!-- 数据导入配置区域 -->
          <div class="config-section">
            <h3>数据导入配置</h3>
            <div class="config-row">
              <div class="config-item">
                <label class="config-label">文件类型</label>
                <t-select v-model="importConfig.fileType" placeholder="选择文件类型" class="config-input">
                  <t-option value="csv" label="CSV" />
                </t-select>
              </div>
              <div class="config-item">
                <label class="config-label">问题选项分隔符</label>
                <t-input
                  v-model="importConfig.optionDelimiter"
                  placeholder="如: \n 或 ,"
                  class="config-input"
                />
              </div>
            </div>
            <div class="config-row">
              <div class="config-item">
                <label class="config-label">文件编码</label>
                <t-select v-model="importConfig.encoding" placeholder="选择文件编码" class="config-input">
                  <t-option value="utf8" label="UTF-8" />
                  <t-option value="gbk" label="GBK" />
                </t-select>
              </div>
              <div class="config-item">
                <label class="config-label">答案分隔符</label>
                <t-input
                  v-model="importConfig.answerDelimiter"
                  placeholder="如: , 或 \n"
                  class="config-input"
                />
              </div>
            </div>
            <div class="config-row">
              <t-button @click="importAnswers" variant="base" class="config-button import-button" id="import-btn">
                导入答案
              </t-button>
            </div>
          </div>

          <!-- 区域选择配置 -->
          <div class="config-section">
            <h3>区域选择配置</h3>
            <div class="area-controls">
              <t-button @click="selectArea" variant="base" class="config-button">
                点击选择区域
              </t-button>
              <t-tag>截图区域: {{ Math.round(screenshotArea.x) }}, {{ Math.round(screenshotArea.y) }}, {{ Math.round(screenshotArea.width) }}x{{ Math.round(screenshotArea.height) }}</t-tag>
            </div>
            <div class="screenshot-preview">
              <div class="preview-images">
                <img v-if="selectedAreaImage" :src="selectedAreaImage" alt="选择区域" class="preview-image" />
                <div v-else class="screenshot-placeholder">
                  <span>暂无截图预览</span>
                </div>
              </div>
            </div>
          </div>

          <!-- 功能区 -->
          <div class="config-section">
            <h3>功能区</h3>
            <div class="config-item">
              <label class="config-label">搜索内容</label>
              <t-textarea
                v-model="ocrResult"
                placeholder="可手动输入搜索内容，或点击下一题进行OCR识别自动搜索"
                :rows="4"
                class="config-input t-textarea"
              />
            </div>
            <div class="action-buttons">
              <t-button @click="searchAnswers" variant="base" class="action-button">
                搜索
              </t-button>
              <t-button @click="nextQuestion" variant="base" class="action-button">
                下一题
              </t-button>
            </div>
          </div>
        </div>
      </div>

      <!-- 可拖动的分隔条 -->
      <div 
        class="resizer" 
        @mousedown="startResize"
        @touchstart="startResize"
      ></div>

      <!-- 右侧答案显示区域 -->
      <div class="answer-panel">
        <div class="answer-content">
          <h2>答案列表</h2>
          
          <!-- 搜索结果 -->
          <div v-if="searchResults.length > 0" class="search-results">
            <div class="search-results-header">
              <h3>搜索结果 ({{ filteredSearchResults.length }}/{{ searchResults.length }}条)</h3>
              <div class="filter-controls" v-if="searchResults.length > 0">
                <t-select 
                  v-model="selectedSearchTypeFilters" 
                  placeholder="筛选题目类型（可多选）"
                  multiple
                  style="width: 300px;"
                  :max="availableSearchTypes.length"
                  @change="handleSearchFilterChange"
                >
                  <t-option 
                    v-for="type in availableSearchTypes" 
                    :key="type" 
                    :value="type" 
                    :label="type" 
                  />
                </t-select>
              </div>
            </div>
            <div class="search-stats">
              <span class="stat-item high-match">高准确率 (≥80%): {{ filteredSearchResults.filter(r => r.score >= 0.8).length }}个</span>
              <span class="stat-item medium-match">中准确率 (50%-79%): {{ filteredSearchResults.filter(r => r.score >= 0.5 && r.score < 0.8).length }}个</span>
              <span class="stat-item low-match">低准确率 (<50%): {{ filteredSearchResults.filter(r => r.score < 0.5).length }}个</span>
            </div>
            <div class="answer-cards">
              <t-card
                v-for="(result, index) in filteredSearchResults"
                :key="index"
                class="answer-card"
              >
                <template #header>
                  <div class="card-header">
                    <t-tag theme="primary">{{ result.item.type }}</t-tag>
                  </div>
                </template>
                <div class="card-content">
                  <div class="card-main">
                    <div class="card-text">
                      <p><strong>题目:</strong> <span v-html="highlightText(result.item.question, result.matches)"></span></p>
                      <div v-if="result.item.options.length > 0">
                        <p><strong>选项:</strong></p>
                        <ul>
                          <li v-for="option in result.item.options" :key="option" v-html="highlightText(option, result.matches)"></li>
                        </ul>
                      </div>
                      <p><strong>答案:</strong></p>
                      <div class="answer-list">
                        <div v-for="(ans, index) in result.item.answer" :key="index" class="answer-item">
                          <span v-html="highlightText(ans, result.matches)"></span>
                        </div>
                      </div>
                      <p><strong>匹配文本:</strong> {{ result.matched }}</p>
                    </div>
                    <div class="match-score-container" :style="getMatchScoreColor(result.score)">
                      <div class="match-score-content">
                        <span class="match-score-label">准确率</span>
                        <span class="match-score-text">{{ (result.score * 100).toFixed(1) }}%</span>
                      </div>
                    </div>
                  </div>
                </div>
              </t-card>
            </div>
          </div>

          <!-- 所有答案 -->
          <div v-else class="all-answers">
            <div class="answers-header">
              <h3>所有答案 ({{ filteredAnswers.length }}/{{ answers.length }}条)</h3>
              <div class="filter-controls" v-if="answers.length > 0">
                <t-select 
                  v-model="selectedTypeFilters" 
                  placeholder="选择题目类型（可多选）"
                  multiple
                  style="width: 300px;"
                  :max="availableTypes.length"
                  @change="handleFilterChange"
                >
                  <t-option 
                    v-for="type in availableTypes" 
                    :key="type" 
                    :value="type" 
                    :label="type" 
                  />
                </t-select>
              </div>
            </div>
            <div v-if="answers.length > 0" class="answer-cards">
              <t-card
                v-for="(answer, index) in filteredAnswers"
                :key="index"
                class="answer-card"
              >
                <template #header>
                  <div class="card-header">
                    <t-tag theme="primary">{{ answer.type }}</t-tag>
                  </div>
                </template>
                <div class="card-content">
                  <p><strong>题目:</strong> {{ answer.question }}</p>
                  <div v-if="answer.options.length > 0">
                    <p><strong>选项:</strong></p>
                    <ul>
                      <li v-for="option in answer.options" :key="option">
                        {{ option }}
                      </li>
                    </ul>
                  </div>
                  <p><strong>答案:</strong></p>
                  <div class="answer-list">
                    <div v-for="(ans, index) in answer.answer" :key="index" class="answer-item">
                      {{ ans }}
                    </div>
                  </div>
                </div>
              </t-card>
            </div>
            
            <!-- 题目类型统计 -->
            <div v-if="answers.length > 0" class="type-stats-inline">
              <span class="stats-label">题目类型统计:</span>
              <span 
                v-for="(count, type) in getFilteredTypeStats()" 
                :key="type" 
                class="type-stat-item-inline"
              >
                {{ type }}: {{ count }}题
              </span>
            </div>
          </div>
          
          <!-- 空状态 -->
          <div v-if="answers.length === 0" class="empty-state">
            <t-empty >
              <template #description>
                <p>请点击左侧的"导入答案"按钮来加载答案数据</p>
              </template>
            </t-empty>
          </div>
        </div>
      </div>
    </div>

    <!-- 区域选择弹窗 -->
    <t-dialog
      v-model:visible="showAreaSelector"
      title="选择截图区域"
      width="98%"
      height="98%"
      :close-on-overlay-click="false"
      :close-on-esc-keydown="false"
      :destroy-on-close="false"
    >
      <div class="area-selector">
        <!-- 上方控制区域 -->
        <div class="area-controls-top">
          <div class="area-info">
            <span v-if="selectedArea" class="info-item">选择区域: {{ selectedArea.x }}, {{ selectedArea.y }} - {{ selectedArea.x + selectedArea.width }} x {{ selectedArea.y + selectedArea.height }}</span>
            <span v-if="selectedArea" class="info-item">缩放比例: {{ getImageScale().x.toFixed(2) }} x {{ getImageScale().y.toFixed(2) }}</span>
            <span v-if="selectedArea" class="info-item">显示区域: {{ Math.round(selectedArea.x / getImageScale().x) }}, {{ Math.round(selectedArea.y / getImageScale().y) }} - {{ Math.round(selectedArea.width / getImageScale().x) }} x {{ Math.round(selectedArea.height / getImageScale().y) }}</span>
          </div>
        </div>
        
        <!-- 截图显示区域 -->
        <div class="screenshot-container">
          <img 
            ref="imageRef"
            :src="fullScreenshot" 
            alt="全屏截图" 
            class="full-screenshot"
            style="width: 100% !important; height: 100% !important; object-fit: fill !important;"
            @mousedown="startSelection"
            @mousemove="updateSelection"
            @mouseup="endSelection"
            @mouseleave="endSelection"
            @dragstart.prevent
            @selectstart.prevent
            @load="imageRef = $event.target"
            draggable="false"
          />
          <div
            v-if="selectedArea && (selectedArea.width > 0 || selectedArea.height > 0)"
            class="selection-box"
            :style="{
              left: (selectedArea.x / getImageScale().x) + 'px',
              top: (selectedArea.y / getImageScale().y) + 'px',
              width: (selectedArea.width / getImageScale().x) + 'px',
              height: (selectedArea.height / getImageScale().y) + 'px'
            }"
          ></div>
        </div>
      </div>
      
      <template #footer>
        <div class="dialog-footer">
          <t-button @click="confirmAreaSelection" variant="base" :disabled="!selectedArea || (selectedArea.width === 0 && selectedArea.height === 0)">
            确定
          </t-button>
          <t-button @click="cancelAreaSelection" variant="base">
            取消
          </t-button>
        </div>
      </template>
    </t-dialog>

    <!-- 错误弹窗 -->
    <t-dialog
      v-model:visible="showErrorDialog"
      :title="errorDialogTitle"
      width="500px"
      :close-on-overlay-click="true"
      :close-on-esc-keydown="true"
    >
      <div class="error-dialog-content">
        <div class="error-message">
          {{ errorDialogContent }}
        </div>
        <div v-if="errorDialogDetails" class="error-details">
          <div class="details-title">详细信息：</div>
          <div class="details-content">
            {{ errorDialogDetails }}
          </div>
        </div>
      </div>
      
      <template #footer>
        <div class="dialog-footer">
          <t-button @click="showErrorDialog = false" variant="base">
            确定
          </t-button>
        </div>
      </template>
    </t-dialog>
  </div>
</template>

<style scoped>
.app-container {
  height: 100vh;
  width: 100vw;
  overflow: hidden;
  background: #ffffff;
  position: relative;
}

.main-layout {
  display: flex;
  height: 100%;
  width: 100%;
  position: relative;
  z-index: 1;
}

.config-panel {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border-right: 1px solid rgba(255, 255, 255, 0.2);
  display: flex;
  flex-direction: column;
  box-shadow: 2px 0 20px rgba(0, 0, 0, 0.1);
}

.config-content {
  padding: 20px;
  height: 100%;
  overflow-y: auto;
}

.config-section {
  margin-top: 10px;
  /* margin-bottom: 2px; */
  padding: 12px;
  background: rgba(255, 255, 255, 0.8);
  border-radius: 12px;
  box-shadow: 0 4px 15px rgba(0, 0, 0, 0.1);
  min-width: 0; /* 防止内容溢出 */
  overflow: hidden; /* 隐藏溢出内容 */
  border: 1px solid rgba(255, 255, 255, 0.3);
  backdrop-filter: blur(5px);
}

.config-section h3 {
  margin: 0 0 2px 0;
  color: #333;
  font-size: 14px;
  font-weight: 600;
}


.local-config {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.config-row {
  display: flex;
  gap: 2px;
  margin-bottom: 2px;
  min-width: 0; /* 防止flex子项溢出 */
  align-items: center; /* 垂直居中对齐 */
}

.config-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0; /* 防止内容溢出 */
  overflow: hidden; /* 隐藏溢出内容 */
}

.config-item-half {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0; /* 防止内容溢出 */
  overflow: hidden; /* 隐藏溢出内容 */
}

/* 答案分隔符容器 - label左对齐 */
.config-item-half:first-child {
  align-items: flex-start; /* label左对齐 */
}

/* 按钮容器 - 与输入框底部对齐 */
.button-container {
  display: flex;
  align-items: flex-end; /* 垂直靠下对齐 */
  height: 100%;
  padding-bottom: 4px; /* 与输入框的margin-bottom对齐 */
}

.config-label {
  font-size: 11px;
  color: #666;
  margin-bottom: 3px;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.config-input {
  margin-bottom: 4px;
  height: 28px;
  font-size: 12px;
  min-width: 0; /* 防止输入框溢出 */
  width: 100%; /* 确保输入框占满容器宽度 */
}

/* 特殊处理textarea */
.config-input.t-textarea {
  height: auto;
  min-height: 80px;
}

.config-button {
  height: 28px;
  font-size: 12px;
  margin-top: 4px;
}

/* 在config-item-half中的按钮样式 */
.config-item-half .config-button {
  margin: 0;
  width: 100%;
  max-width: 120px; /* 限制按钮最大宽度 */
}

/* 按钮容器中的按钮样式 */
.button-container .config-button {
  height: 28px; /* 与输入框相同高度 */
  width: 100%;
  max-width: none; /* 移除最大宽度限制，允许拉伸 */
}

.area-controls {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 10px;
  margin-bottom: 10px;
}

.area-controls .config-button {
  margin: 0;
  flex-shrink: 0;
  height: 24px;
  padding: 0 8px;
  font-size: 11px;
}

.area-controls .t-tag {
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.import-button {
  width: 100%;
}

.screenshot-preview {
  margin-top: 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
  overflow: hidden;
  height: 120px;
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f8f8f8;
}

.preview-images {
  display: flex;
  gap: 10px;
  width: 100%;
  height: 100%;
  /* overflow: hidden; */
}

.preview-image {
  width: 100%; /* 占满容器宽度 */
  height: 100%;
  /* object-fit: contain;  保持比例，完整显示 */
  border-radius: 4px;
  border: 1px solid #e0e0e0;
}

.screenshot-placeholder {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100%;
  height: 100%;
  background-color: #f0f0f0;
  border-radius: 4px;
  border: 1px dashed #ccc;
}

.screenshot-placeholder span {
  color: #999;
  font-size: 14px;
}

.action-buttons {
  display: flex;
  gap: 8px;
  margin-top: 4px;
}

.action-button {
  flex: 1;
}

.answer-panel {
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(10px);
  flex: 1; /* Allow answer panel to grow and take available space */
  display: flex;
  flex-direction: column;
  box-shadow: -2px 0 20px rgba(0, 0, 0, 0.1);
}

.answer-content {
  padding: 20px;
  height: 100%;
  overflow-y: auto;
  background: rgba(255, 255, 255, 0.7);
  backdrop-filter: blur(5px);
  border-radius: 8px;
  margin: 10px;
}

.answer-content h2 {
  margin: 0 0 20px 0;
  color: #2c3e50;
  font-weight: 700;
  font-size: 24px;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
}

.answer-content h3 {
  margin: 20px 0 15px 0;
  color: #34495e;
  font-weight: 600;
  font-size: 18px;
}

.answer-cards {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.answer-card {
  transition: all 0.3s ease;
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(5px);
  border: 1px solid rgba(255, 255, 255, 0.3);
  box-shadow: 0 4px 15px rgba(0, 0, 0, 0.1);
}

.answer-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.15);
  background: rgba(255, 255, 255, 1);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.match-score {
  font-weight: bold;
  color: #0052d9;
}

.card-content p {
  margin: 8px 0;
}

.card-content ul {
  margin: 8px 0;
}

.answer-list {
  margin: 8px 0;
}

.answer-item {
  padding: 8px 12px;
  margin: 4px 0;
  background-color: #f8f9fa;
  border-left: 3px solid #0052d9;
  border-radius: 4px;
  font-size: 14px;
  line-height: 1.4;
}

.answers-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-wrap: wrap;
  gap: 16px;
}

.search-results-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-wrap: wrap;
  gap: 16px;
}

.filter-controls {
  display: flex;
  align-items: center;
  gap: 12px;
}

.filter-controls .t-select {
  min-width: 300px;
}

.filter-controls .t-select__tags {
  max-width: 280px;
  overflow: hidden;
}

.card-content ul {
  margin: 8px 0;
  padding-left: 20px;
}

.card-content li {
  margin: 4px 0;
}

.area-selector {
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
  max-height: 100%;
  overflow: hidden;
}

.area-controls-top {
  display: flex;
  justify-content: flex-start;
  align-items: center;
  padding: 8px 0;
  flex-shrink: 0; /* 防止压缩 */
}

.area-info {
  margin: 0;
  font-size: 12px;
  color: #666;
  flex: 1;
  line-height: 1.4;
  display: flex;
  gap: 20px;
  align-items: center;
}

.info-item {
  white-space: nowrap;
  padding: 6px 12px;
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(5px);
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.3);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  font-weight: 500;
  color: #2c3e50;
}

.screenshot-container {
  position: relative;
  flex: 1;
  overflow: hidden;
  background: #f0f0f0;
  border-radius: 8px;
  margin-bottom: 16px;
  height: calc(100vh - 200px);
  min-height: 500px;
  display: flex;
  align-items: stretch;
  justify-content: stretch;
  width: 100%;
}

.full-screenshot {
  width: 100% !important;
  height: 100% !important;
  cursor: crosshair;
  user-select: none;
  -webkit-user-select: none;
  -moz-user-select: none;
  -ms-user-select: none;
  pointer-events: auto;
  -webkit-user-drag: none;
  -khtml-user-drag: none;
  -moz-user-drag: none;
  -o-user-drag: none;
  user-drag: none;
  object-fit: fill !important; /* 强制拉伸填满容器 */
  border-radius: 8px;
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  margin: 0;
  padding: 0;
}

.selection-box {
  position: absolute;
  border: 2px solid #1890ff;
  background: rgba(24, 144, 255, 0.1);
  pointer-events: none;
  z-index: 10;
  border-radius: 4px;
}

.search-results {
  margin-bottom: 30px;
}

.search-stats {
  display: flex;
  gap: 16px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.stat-item {
  padding: 4px 12px;
  border-radius: 16px;
  font-size: 12px;
  font-weight: 500;
}

.stat-item.high-match {
  background: #f6ffed;
  color: #52c41a;
  border: 1px solid #b7eb8f;
}

.stat-item.medium-match {
  background: #fff7e6;
  color: #fa8c16;
  border: 1px solid #ffd591;
}

.stat-item.low-match {
  background: #fff2f0;
  color: #ff4d4f;
  border: 1px solid #ffccc7;
}

/* 题目类型统计样式 */
.type-stats {
  margin-bottom: 20px;
  padding: 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 12px;
  border: none;
  box-shadow: 0 4px 15px rgba(102, 126, 234, 0.2);
  position: relative;
  overflow: hidden;
}

.type-stats::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.1) 0%, rgba(255, 255, 255, 0.05) 100%);
  pointer-events: none;
}

.stats-title {
  font-size: 16px;
  font-weight: 700;
  color: #ffffff;
  margin-bottom: 16px;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
  position: relative;
  z-index: 1;
}

.stats-content {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  position: relative;
  z-index: 1;
}

.type-stat-item {
  padding: 8px 16px;
  background: rgba(255, 255, 255, 0.95);
  border: none;
  border-radius: 20px;
  font-size: 13px;
  font-weight: 600;
  color: #4a5568;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}

.type-stat-item::before {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.4), transparent);
  transition: left 0.5s ease;
}

.type-stat-item:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
  background: rgba(255, 255, 255, 1);
}

.type-stat-item:hover::before {
  left: 100%;
}

/* 内联题目类型统计样式 */
.type-stats-inline {
  margin-top: 16px;
  padding: 12px 16px;
  background: rgba(255, 255, 255, 0.8);
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.3);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.stats-label {
  font-size: 14px;
  font-weight: 600;
  color: #2c3e50;
  white-space: nowrap;
}

.type-stat-item-inline {
  padding: 4px 10px;
  background: rgba(255, 255, 255, 0.9);
  border: 1px solid rgba(0, 0, 0, 0.1);
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
  color: #495057;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  white-space: nowrap;
}

/* OCR配置样式 */
.ocr-mode-status {
  display: flex;
  align-items: center;
  align-content: center;
  justify-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 4px;
  margin-top: 4px;
}


.ocr-mode-status > * {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  max-height: 100%;
}



.ocr-mode-status .connection-status {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
}



/* 强制垂直居中 */
.t-radio-group {
  display: flex !important;
  align-items: center !important;
  justify-content: center !important;
}

.t-radio-group .t-radio-group__group {
  display: flex !important;
  align-items: center !important;
  gap: 16px !important;
}

.t-radio {
  display: flex !important;
  align-items: center !important;
  margin: 0 !important;
  padding: 0 8px !important;
}

.t-radio__label {
  display: flex !important;
  align-items: center !important;
  line-height: 1 !important;
  font-size: 14px !important;
  font-weight: 500 !important;
}

/* mode-selector 样式 */
.mode-selector {
  display: flex;
  align-items: center;
  justify-content: center;
}

.connection-status {
  display: flex;
  align-items: center;
  height: 100%;
  flex-shrink: 0; /* 防止收缩 */
}

.connection-status-right {
  display: flex;
  align-items: center;
  height: 100%;
  flex-shrink: 0; /* 防止收缩 */
  margin-left: auto; /* 靠右侧显示 */
}

.status-tag {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 12px;
  font-weight: 500;
  border: none;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 20px;
}



.status-success {
  background: linear-gradient(135deg, #a8e6cf, #88d8c0);
  color: #2d5a3d;
}

.status-error {
  background: linear-gradient(135deg, #ffb3ba, #ff8a95);
  color: #8b2635;
}

.status-connecting {
  background: linear-gradient(135deg, #fff3cd, #ffeaa7);
  color: #856404;
}

.status-default {
  background: linear-gradient(135deg, #e8f4fd, #d1e7dd);
  color: #6c757d;
}

.empty-state {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 400px;
  padding: 40px;
  background: rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  border: 1px solid rgba(255, 255, 255, 0.3);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
}

.resizer {
  width: 6px;
  background: transparent;
  cursor: col-resize;
  position: relative;
  transition: all 0.2s ease;
  margin: 0 1px;
}

.resizer:hover {
  background: transparent;
}

.resizer:active {
  background: transparent;
}

.resizer::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 1px;
  height: 20px;
  background: transparent;
  border-radius: 1px;
}

.resizer::after {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 1px;
  height: 10px;
  background: transparent;
  border-radius: 1px;
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
  padding: 0;
}

.dialog-buttons {
  display: flex;
  gap: 8px;
  align-items: center;
}

.dialog-buttons .t-button {
  height: 24px;
  padding: 0 8px;
  font-size: 12px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 0;
}

.dialog-footer .t-button {
  min-width: 80px;
}

.error-dialog-content {
  padding: 20px 0;
}

.error-message {
  font-size: 14px;
  color: #333;
  margin-bottom: 15px;
  line-height: 1.5;
}

.error-details {
  background: #f8f9fa;
  border: 1px solid #e9ecef;
  border-radius: 6px;
  padding: 15px;
  margin-top: 15px;
}

.details-title {
  font-weight: 600;
  color: #495057;
  margin-bottom: 8px;
  font-size: 13px;
}

.details-content {
  font-size: 12px;
  color: #6c757d;
  line-height: 1.6;
  white-space: pre-line;
}

/* 卡片内容布局 */
.card-main {
  display: flex;
  gap: 16px;
  align-items: flex-start;
}

.card-text {
  flex: 1;
}

.match-score-container {
  min-width: 80px;
  height: 80px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px solid;
  flex-shrink: 0;
}

.match-score-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
}

.match-score-label {
  font-size: 12px;
  font-weight: normal;
  margin-bottom: 4px;
  opacity: 0.8;
}

.match-score-text {
  font-size: 18px;
  font-weight: bold;
  text-align: center;
}


</style>
