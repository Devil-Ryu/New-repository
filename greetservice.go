package main

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"

	"github.com/kbinani/screenshot"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// ExamService 考试助手服务
type ExamService struct{}

// OCRConfig OCR配置
type OCRConfig struct {
	Mode   string `json:"mode"`   // "online" 或 "local"
	URL    string `json:"url"`    // 在线OCR URL
	APIKey string `json:"apiKey"` // API密钥
	Status string `json:"status"` // 连接状态
}

// ImportConfig 导入配置
type ImportConfig struct {
	FileType  string `json:"fileType"`  // "excel" 或 "csv"
	Encoding  string `json:"encoding"`  // 文件编码
	Delimiter string `json:"delimiter"` // 答案分隔符
}

// ScreenshotArea 截图区域
type ScreenshotArea struct {
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Image  string `json:"image"` // base64编码的图片
}

// AnswerItem 答案项
type AnswerItem struct {
	Type     string   `json:"type"`     // 题目类型
	Question string   `json:"question"` // 题目内容
	Options  []string `json:"options"`  // 选项
	Answer   []string `json:"answer"`   // 答案
}

// 校验过程可能返回类型
type HeaderError struct {
	Missing []string // 缺失字段
	Extra   []string // 多余字段
}

func (e HeaderError) Error() string {
	msgs := []string{}
	if len(e.Missing) > 0 {
		msgs = append(msgs, fmt.Sprintf("缺失字段：%v", e.Missing))
	}
	if len(e.Extra) > 0 {
		msgs = append(msgs, fmt.Sprintf("多余字段：%v", e.Extra))
	}
	return strings.Join(msgs, "; ")
}

// SearchResult 搜索结果
type SearchResult struct {
	Item    AnswerItem `json:"item"`
	Score   float64    `json:"score"`   // 匹配度
	Matched string     `json:"matched"` // 匹配的文本
	Matches []int      `json:"matches"` // 匹配位置
}

// FileDialogResult 文件对话框结果
type FileDialogResult struct {
	FilePath string `json:"filePath"`
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
}

// OCRResult OCR识别结果
type OCRResult struct {
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
	BBox       struct {
		XMin   int     `json:"xmin"`
		YMin   int     `json:"ymin"`
		XMax   int     `json:"xmax"`
		YMax   int     `json:"ymax"`
		Points [][]int `json:"points"`
	} `json:"bbox"`
}

// OCRResponse OCR响应结构
type OCRResponse struct {
	Success bool `json:"success"`
	Data    struct {
		TextCount int         `json:"text_count"`
		Results   []OCRResult `json:"results"`
	} `json:"data"`
}

// OCRService OCR服务结构
type OCRService struct {
	ServerURL string
	Client    *http.Client
}

// 全局OCR服务实例
// 全局OCR服务变量（已废弃，保留用于兼容性）

// ProcessImage 处理图片进行OCR识别
func (o *OCRService) ProcessImage(imageData []byte) ([]OCRResult, error) {
	// 将图片数据编码为base64
	base64Data := base64.StdEncoding.EncodeToString(imageData)

	// 准备请求数据
	requestData := map[string]string{
		"image": base64Data,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("编码请求数据失败: %v", err)
	}

	// 发送HTTP请求
	req, err := http.NewRequest("POST", o.ServerURL+"/ocr", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建OCR请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送OCR请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取OCR响应失败: %v", err)
	}

	// 解析JSON响应
	var ocrResp OCRResponse
	err = json.Unmarshal(body, &ocrResp)
	if err != nil {
		return nil, fmt.Errorf("解析OCR响应失败: %v", err)
	}

	if !ocrResp.Success {
		return nil, fmt.Errorf("OCR服务返回错误")
	}

	return ocrResp.Data.Results, nil
}

// OpenFileDialog 打开文件对话框
func (e *ExamService) OpenFileDialog(title string, fileType string) (FileDialogResult, error) {
	// 使用Wails v3的文件对话框API
	dialog := application.OpenFileDialog()

	// 设置标题
	dialog.SetTitle(title)

	// 设置文件过滤器
	if fileType == "csv" {
		dialog.AddFilter("CSV文件", "*.csv")
	} else if fileType == "excel" {
		dialog.AddFilter("Excel文件", "*.xlsx;*.xls")
	}

	// 允许选择所有文件类型
	dialog.AddFilter("所有文件", "*.*")

	// 确保可以选择文件
	dialog.CanChooseFiles(true)
	dialog.CanChooseDirectories(false)

	// 尝试附加到主窗口（如果可用）
	app := application.Get()
	if app != nil {
		windows := app.Window.GetAll()
		if len(windows) > 0 {
			mainWindow := windows[0]
			dialog.AttachToWindow(mainWindow)
		}
	}

	// 提示用户选择单个文件
	filePath, err := dialog.PromptForSingleSelection()
	if err != nil {
		return FileDialogResult{
			FilePath: "",
			Success:  false,
			Error:    fmt.Sprintf("打开文件对话框失败: %v", err),
		}, nil
	}

	// 如果用户取消了选择，filePath为空
	if filePath == "" {
		return FileDialogResult{
			FilePath: "",
			Success:  false,
			Error:    "用户取消了文件选择",
		}, nil
	}

	return FileDialogResult{
		FilePath: filePath,
		Success:  true,
	}, nil
}

// ReadFileContent 读取文件内容
// getEncoding 根据编码名称获取对应的编码器
func getEncoding(encodingName string) (encoding.Encoding, error) {
	switch strings.ToLower(encodingName) {
	case "utf8", "utf-8":
		return nil, nil // UTF-8是默认编码
	case "gbk", "gb2312":
		return simplifiedchinese.GBK, nil
	default:
		return nil, fmt.Errorf("不支持的编码格式: %s，仅支持UTF-8和GBK", encodingName)
	}
}

func (e *ExamService) ReadFileContent(filePath string, encoding string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("无法打开文件: %v", err)
	}
	defer file.Close()

	var reader io.Reader = file

	// 根据编码处理文件
	if encoding != "utf8" && encoding != "utf-8" {
		enc, err := getEncoding(encoding)
		if err != nil {
			return "", fmt.Errorf("编码设置错误: %v", err)
		}

		if enc != nil {
			reader = transform.NewReader(file, enc.NewDecoder())
		}
	}

	// 读取文件内容
	content, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %v", err)
	}

	return string(content), nil
}

// ParseCSVFile 解析CSV文件
func (e *ExamService) ParseCSVFile(filePath string, encoding string, optionSeparator string, answerSeparator string) ([]AnswerItem, error) {
	var answers []AnswerItem

	// 打开文件
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("无法打开文件: %v", err)
	}
	defer f.Close()

	// 解码器处理
	var reader io.Reader
	switch strings.ToLower(encoding) {
	case "gbk", "gb2312":
		reader = transform.NewReader(f, simplifiedchinese.GBK.NewDecoder())
	case "utf-8", "utf8":
		reader = f
	default:
		return nil, fmt.Errorf("不支持的编码格式: %s", encoding)
	}

	csvReader := csv.NewReader(reader)
	csvReader.TrimLeadingSpace = true

	// 读取标题行
	headers, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("读取标题行失败: %v", err)
	}

	expected := map[string]int{"类型": -1, "题目": -1, "选项": -1, "答案": -1}
	for i, h := range headers {
		if _, ok := expected[h]; ok {
			expected[h] = i
		}
	}

	// 检查缺失字段
	var missing []string
	for key, idx := range expected {
		if idx == -1 {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("缺少字段: %s", strings.Join(missing, ", "))
	}

	// 读取数据行
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("读取数据失败: %v", err)
		}

		answer := AnswerItem{
			Type:     strings.TrimSpace(record[expected["类型"]]),
			Question: strings.TrimSpace(record[expected["题目"]]),
			Options:  []string{},
			Answer:   []string{},
		}

		// 拆分选项
		optionsStr := record[expected["选项"]]
		if optionSeparator != "" {
			separator := e.parseSeparator(optionSeparator)
			answer.Options = strings.Split(optionsStr, separator)
		} else {
			answer.Options = []string{optionsStr}
		}

		// 拆分答案
		answerStr := record[expected["答案"]]
		if answerStr != "" {
			separator := e.parseSeparator(answerSeparator)
			answer.Answer = strings.Split(answerStr, separator)
		}

		answers = append(answers, answer)
	}

	return answers, nil
}

// parseSeparator 解析分隔符，支持转义字符
func (e *ExamService) parseSeparator(separator string) string {
	switch separator {
	case "\\n":
		return "\n"
	case "\\t":
		return "\t"
	case "\\r":
		return "\r"
	case "\\s":
		return " "
	default:
		return separator
	}
}

// TestOCRConnection 测试OCR连接
func (e *ExamService) TestOCRConnection(config OCRConfig) (string, error) {
	if config.URL == "" {
		return "连接失败", fmt.Errorf("未配置OCR服务URL")
	}

	// 构建健康检查URL
	healthURL := config.URL
	if !strings.HasSuffix(healthURL, "/") {
		healthURL += "/"
	}
	healthURL += "health"

	// 发送HTTP请求到健康检查端点
	req, err := http.NewRequest("GET", healthURL, nil)
	if err != nil {
		return "连接失败", fmt.Errorf("创建健康检查请求失败: %v", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "连接失败", fmt.Errorf("健康检查请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "连接失败", fmt.Errorf("读取健康检查响应失败: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != 200 {
		return "连接失败", fmt.Errorf("健康检查失败，状态码: %d", resp.StatusCode)
	}

	// 尝试解析JSON响应
	var healthResp struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(body, &healthResp); err == nil {
		if healthResp.Success {
			return "连接成功", nil
		} else {
			return "连接失败", fmt.Errorf("OCR服务报告错误: %s", healthResp.Message)
		}
	}

	// 如果无法解析JSON，但HTTP状态码是200，也认为连接成功
	return "连接成功", nil
}

// TestLocalOCR 测试本地OCR功能
func (e *ExamService) TestLocalOCR() (string, error) {
	// 使用默认的本地OCR服务URL
	defaultURL := "http://127.0.0.1:8080"

	// 读取test.png文件
	imagePath := "test.png"
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return "", fmt.Errorf("读取图像文件失败: %v", err)
	}

	// 使用新的OCR服务处理图像
	result, err := e.performOCRWithURL(imageData, defaultURL)
	if err != nil {
		return "", fmt.Errorf("OCR处理失败: %v", err)
	}

	if result == "" {
		return "未检测到任何文本内容", nil
	}

	return fmt.Sprintf("OCR处理完成，识别结果：\n%s", result), nil
}

// TakeScreenshot 截取屏幕
func (e *ExamService) TakeScreenshot() (string, error) {
	// 获取主显示器的边界
	bounds := screenshot.GetDisplayBounds(0)

	// 截取屏幕
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return "", fmt.Errorf("截图失败: %v", err)
	}

	// 将图片编码为PNG格式
	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	if err != nil {
		return "", fmt.Errorf("图片编码失败: %v", err)
	}

	// 转换为base64编码
	base64Data := base64.StdEncoding.EncodeToString(buf.Bytes())

	// 返回data URL格式
	return "data:image/png;base64," + base64Data, nil
}

// TakeScreenshotWithWindowControl 带窗口控制的截图
func (e *ExamService) TakeScreenshotWithWindowControl() (string, error) {
	// 获取应用实例
	app := application.Get()
	if app == nil {
		return "", fmt.Errorf("无法获取应用实例")
	}

	// 获取所有窗口
	windows := app.Window.GetAll()
	if len(windows) == 0 {
		return "", fmt.Errorf("没有找到窗口")
	}

	window := windows[0]

	// 1. 隐藏窗口
	window.Minimise()

	// 2. 等待一小段时间确保窗口完全隐藏
	time.Sleep(500 * time.Millisecond)

	// 3. 截取屏幕
	screenshot, err := e.TakeScreenshot()
	if err != nil {
		// 即使截图失败也要恢复窗口
		window.Restore()
		return "", err
	}

	// 4. 恢复窗口
	window.Restore()

	return screenshot, nil
}

// SelectArea 选择截图区域
func (e *ExamService) SelectArea(screenshotData string) (ScreenshotArea, error) {
	// 这个函数现在主要用于接收前端已经裁剪好的图片
	// 前端会直接传递裁剪后的图片数据
	area := ScreenshotArea{
		X:      0, // 裁剪后的图片，坐标从0开始
		Y:      0,
		Width:  0, // 宽度和高度会在前端设置
		Height: 0,
		Image:  screenshotData, // 这里应该是裁剪后的图片
	}
	return area, nil
}

// PerformOCR 执行OCR识别
func (e *ExamService) PerformOCR(area ScreenshotArea, config OCRConfig) (string, error) {
	if area.Image == "" {
		return "", fmt.Errorf("没有截图数据")
	}

	// 解码base64图片数据
	imageData, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(area.Image, "data:image/png;base64,"))
	if err != nil {
		return "", fmt.Errorf("图片解码失败: %v", err)
	}

	// 解码PNG图片
	img, err := png.Decode(bytes.NewReader(imageData))
	if err != nil {
		return "", fmt.Errorf("图片解码失败: %v", err)
	}

	// 如果指定了区域，裁剪图片
	if area.Width > 0 && area.Height > 0 {
		bounds := img.Bounds()
		if area.X+area.Width <= bounds.Dx() && area.Y+area.Height <= bounds.Dy() {
			img = img.(interface {
				SubImage(r image.Rectangle) image.Image
			}).SubImage(image.Rect(area.X, area.Y, area.X+area.Width, area.Y+area.Height))
		}
	}

	// 重新编码为PNG
	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	if err != nil {
		return "", fmt.Errorf("图片编码失败: %v", err)
	}

	// 使用配置的OCR服务URL进行识别
	if config.URL != "" {
		return e.performOCRWithURL(buf.Bytes(), config.URL)
	}

	// 如果没有配置OCR URL，返回模拟结果
	return "这是一个模拟的OCR识别结果", nil
}

// performOCRWithURL 使用指定URL的OCR服务进行识别
func (e *ExamService) performOCRWithURL(imageData []byte, serverURL string) (string, error) {
	// 将图片数据编码为base64
	base64Data := base64.StdEncoding.EncodeToString(imageData)

	// 准备请求数据
	requestData := map[string]string{
		"image": base64Data,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return "", fmt.Errorf("编码请求数据失败: %v", err)
	}

	// 构建OCR请求URL
	ocrURL := serverURL
	if !strings.HasSuffix(ocrURL, "/") {
		ocrURL += "/"
	}
	ocrURL += "ocr"

	// 发送HTTP请求
	req, err := http.NewRequest("POST", ocrURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建OCR请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送OCR请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取OCR响应失败: %v", err)
	}

	// 解析JSON响应
	var ocrResp OCRResponse
	err = json.Unmarshal(body, &ocrResp)
	if err != nil {
		return "", fmt.Errorf("解析OCR响应失败: %v", err)
	}

	if !ocrResp.Success {
		return "", fmt.Errorf("OCR服务返回错误")
	}

	// 只保留文字内容，合并所有识别结果
	var allText strings.Builder
	for i, result := range ocrResp.Data.Results {
		if i > 0 {
			allText.WriteString(" ")
		}
		allText.WriteString(strings.TrimSpace(result.Text))
	}

	return allText.String(), nil
}

// performOnlineOCR 使用在线OCR服务
func (e *ExamService) performOnlineOCR(imageData []byte, config OCRConfig) (string, error) {
	// 准备表单数据
	formData := url.Values{}
	formData.Set("apikey", config.APIKey)
	formData.Set("language", "chs")
	formData.Set("isOverlayRequired", "false")
	formData.Set("filetype", "png")
	formData.Set("detectOrientation", "true")

	// 创建multipart表单
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加文件
	part, err := writer.CreateFormFile("file", "screenshot.png")
	if err != nil {
		return "", fmt.Errorf("创建表单失败: %v", err)
	}
	_, err = part.Write(imageData)
	if err != nil {
		return "", fmt.Errorf("写入图片数据失败: %v", err)
	}

	// 添加其他参数
	for key, values := range formData {
		for _, value := range values {
			err := writer.WriteField(key, value)
			if err != nil {
				return "", fmt.Errorf("写入表单字段失败: %v", err)
			}
		}
	}

	err = writer.Close()
	if err != nil {
		return "", fmt.Errorf("关闭表单失败: %v", err)
	}

	// 发送HTTP请求
	req, err := http.NewRequest("POST", config.URL, &buf)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析JSON响应
	var result struct {
		ParsedResults []struct {
			ParsedText string `json:"ParsedText"`
		} `json:"ParsedResults"`
		ErrorMessage string `json:"ErrorMessage"`
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	if result.ErrorMessage != "" {
		return "", fmt.Errorf("OCR服务错误: %s", result.ErrorMessage)
	}

	if len(result.ParsedResults) == 0 {
		return "", fmt.Errorf("没有识别到文本")
	}

	return strings.TrimSpace(result.ParsedResults[0].ParsedText), nil
}

// SearchAnswers 搜索答案
func (e *ExamService) SearchAnswers(answers []AnswerItem, query string) ([]SearchResult, error) {
	log.Println("SearchAnswers", answers, query)
	results := []SearchResult{}
	query = strings.ToLower(strings.TrimSpace(query))

	if query == "" {
		return results, nil
	}

	for _, answer := range answers {
		question := answer.Question
		questionLower := strings.ToLower(question)
		score := 0.0
		matched := ""
		maxScore := 0.0
		bestMatches := []int{}

		// 计算题目重合度（使用原始文本计算匹配位置）
		questionScore, _ := e.calculateOverlapScore(query, questionLower)
		questionMatches := e.calculateMatchesForOriginalText(question, query)
		if questionScore > maxScore {
			maxScore = questionScore
			matched = query
			bestMatches = questionMatches
		}

		// 计算答案重合度
		for _, ans := range answer.Answer {
			ansLower := strings.ToLower(ans)
			ansScore, _ := e.calculateOverlapScore(query, ansLower)
			ansMatches := e.calculateMatchesForOriginalText(ans, query)
			if ansScore > maxScore {
				maxScore = ansScore
				matched = query
				bestMatches = ansMatches
			}
		}

		// 计算选项重合度
		for _, option := range answer.Options {
			optionLower := strings.ToLower(option)
			optionScore, _ := e.calculateOverlapScore(query, optionLower)
			optionMatches := e.calculateMatchesForOriginalText(option, query)
			optionScore = optionScore * 0.8 // 选项权重稍低
			if optionScore > maxScore {
				maxScore = optionScore
				matched = query
				bestMatches = optionMatches
			}
		}

		score = maxScore

		if score > 0 {
			// 限制分数不超过1.0
			if score > 1.0 {
				score = 1.0
			}
			log.Printf("搜索结果: 题目='%s', 分数=%.2f, 匹配位置=%v", answer.Question, score, bestMatches)
			results = append(results, SearchResult{
				Item:    answer,
				Score:   score,
				Matched: matched,
				Matches: bestMatches,
			})
		}
	}

	return results, nil
}

// calculateOverlapScore 计算重合度分数
func (e *ExamService) calculateOverlapScore(query, text string) (float64, []int) {
	if query == "" || text == "" {
		return 0.0, nil
	}

	// 完全匹配
	if query == text {
		return 1.0, []int{0, len(query)}
	}

	// 连续包含匹配
	if strings.Contains(text, query) {
		start := strings.Index(text, query)
		queryLen := len(query)
		textLen := len(text)
		ratio := float64(queryLen) / float64(textLen)

		// 生成匹配位置的字符索引
		matches := []int{}
		for i := start; i < start+queryLen; i++ {
			matches = append(matches, i)
		}

		return 0.8 + (ratio * 0.2), matches
	}

	// 模糊匹配和不连续匹配
	return e.fuzzyMatch(query, text)
}

// fuzzyMatch 模糊匹配算法
func (e *ExamService) fuzzyMatch(query, text string) (float64, []int) {
	queryRunes := []rune(query)
	textRunes := []rune(text)

	if len(queryRunes) == 0 || len(textRunes) == 0 {
		return 0.0, nil
	}

	// 动态规划计算最长公共子序列
	dp := make([][]int, len(queryRunes)+1)
	for i := range dp {
		dp[i] = make([]int, len(textRunes)+1)
	}

	// 填充DP表
	for i := 1; i <= len(queryRunes); i++ {
		for j := 1; j <= len(textRunes); j++ {
			if queryRunes[i-1] == textRunes[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max(dp[i-1][j], dp[i][j-1])
			}
		}
	}

	lcsLength := dp[len(queryRunes)][len(textRunes)]
	if lcsLength == 0 {
		return 0.0, nil
	}

	// 回溯找到匹配位置
	matches := e.backtrackMatches(dp, queryRunes, textRunes)

	// 计算分数：LCS长度 / 查询长度
	score := float64(lcsLength) / float64(len(queryRunes))

	// 根据匹配的连续性调整分数
	continuityBonus := e.calculateContinuityBonus(matches)
	score = score * (0.6 + continuityBonus*0.4)

	return score, matches
}

// backtrackMatches 回溯找到匹配位置
func (e *ExamService) backtrackMatches(dp [][]int, query, text []rune) []int {
	matches := []int{}
	i, j := len(query), len(text)

	for i > 0 && j > 0 {
		if query[i-1] == text[j-1] {
			matches = append([]int{j - 1}, matches...)
			i--
			j--
		} else if dp[i-1][j] > dp[i][j-1] {
			i--
		} else {
			j--
		}
	}

	// 转换为字符位置（考虑中文字符）
	textStr := string(text)
	charMatches := []int{}
	charIndex := 0

	for i := range textStr {
		if contains(matches, charIndex) {
			charMatches = append(charMatches, i)
		}
		charIndex++
	}

	return charMatches
}

// contains 检查切片是否包含某个值
func contains(slice []int, value int) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// calculateContinuityBonus 计算连续性奖励
func (e *ExamService) calculateContinuityBonus(matches []int) float64 {
	if len(matches) <= 1 {
		return 0.0
	}

	continuous := 0
	total := len(matches)

	for i := 1; i < len(matches); i++ {
		if matches[i] == matches[i-1]+1 {
			continuous++
		}
	}

	return float64(continuous) / float64(total-1)
}

// max 返回两个整数中的较大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// mapMatchesToOriginalText 将小写文本的匹配位置映射到原始文本
func (e *ExamService) mapMatchesToOriginalText(original, lower string, matches []int) []int {
	if len(matches) == 0 {
		return matches
	}

	log.Printf("映射匹配位置: 原始文本='%s', 小写文本='%s', 匹配位置=%v", original, lower, matches)

	// 对于简单的ASCII文本，位置通常是相同的
	// 但对于包含大写字母的文本，需要调整
	originalMatches := []int{}

	for _, match := range matches {
		if match < len(lower) {
			// 找到在原始文本中对应的位置
			originalPos := e.findCorrespondingPosition(original, lower, match)
			if originalPos >= 0 {
				originalMatches = append(originalMatches, originalPos)
			}
		}
	}

	log.Printf("映射后的匹配位置: %v", originalMatches)
	return originalMatches
}

// findCorrespondingPosition 找到原始文本中对应的位置
func (e *ExamService) findCorrespondingPosition(original, lower string, lowerPos int) int {
	if lowerPos >= len(lower) {
		return -1
	}

	// 对于包含大写字母的情况，需要找到对应的原始位置
	// 例如：原始="Vue.js"，小写="vue.js"，小写位置2对应原始位置0
	lowerPosCount := 0

	for i := range original {
		if lowerPosCount == lowerPos {
			return i
		}
		lowerPosCount++
	}

	// 如果没找到，返回-1
	return -1
}

// 简化匹配位置计算，直接基于原始文本计算
func (e *ExamService) calculateMatchesForOriginalText(originalText, query string) []int {
	matches := []int{}
	originalLower := strings.ToLower(originalText)
	queryLower := strings.ToLower(query)

	// 简单的包含匹配
	if strings.Contains(originalLower, queryLower) {
		start := strings.Index(originalLower, queryLower)
		for i := start; i < start+len(queryLower); i++ {
			matches = append(matches, i)
		}
	}

	return matches
}

// NextQuestion 下一题功能
func (e *ExamService) NextQuestion(area ScreenshotArea, config OCRConfig) (string, error) {
	// 1. 重新截图
	screenshot, err := e.TakeScreenshot()
	if err != nil {
		return "", err
	}

	// 2. 执行OCR识别
	area.Image = screenshot
	ocrResult, err := e.PerformOCR(area, config)
	if err != nil {
		return "", err
	}

	return ocrResult, nil
}

// HideWindow 隐藏应用窗口
func (e *ExamService) HideWindow() error {
	// 获取应用实例
	app := application.Get()
	if app == nil {
		return fmt.Errorf("无法获取应用实例")
	}

	// 获取所有窗口
	windows := app.Window.GetAll()
	if len(windows) == 0 {
		return fmt.Errorf("没有找到窗口")
	}

	// 隐藏第一个窗口（主窗口）
	// 使用更安全的方式隐藏窗口
	windows[0].Minimise()
	return nil
}

// ShowWindow 显示应用窗口
func (e *ExamService) ShowWindow() error {
	// 获取应用实例
	app := application.Get()
	if app == nil {
		return fmt.Errorf("无法获取应用实例")
	}

	// 获取所有窗口
	windows := app.Window.GetAll()
	if len(windows) == 0 {
		return fmt.Errorf("没有找到窗口")
	}

	// 显示第一个窗口（主窗口）
	// 使用更安全的方式显示窗口
	windows[0].Restore()
	return nil
}
