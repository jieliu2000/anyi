# 多模态应用教程

本教程将指导您如何使用 Anyi 构建处理文本和图像的多模态 AI 应用程序。

## 多模态概述

多模态 AI 应用能够同时处理多种类型的输入，如：

- 文本和图像
- 音频和文本
- 视频和文本

Anyi 目前主要支持文本和图像的多模态处理。

## 支持的提供商

目前支持多模态的 LLM 提供商：

| 提供商       | 支持的模态  | 主要模型        |
| ------------ | ----------- | --------------- |
| OpenAI       | 文本 + 图像 | GPT-4V, GPT-4o  |
| Azure OpenAI | 文本 + 图像 | GPT-4V          |
| Ollama       | 文本 + 图像 | LLaVA, Bakllava |
| 通义千问     | 文本 + 图像 | Qwen-VL         |

## 基本图像处理

### 1. 图像描述生成

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/dashscope"
    "github.com/jieliu2000/anyi/llm/chat"
)

func main() {
    // 创建支持视觉的通义千问客户端
    config := dashscope.DefaultConfig(os.Getenv("DASHSCOPE_API_KEY"))
    config.Model = "qwen-vl-max" // 使用支持图像的模型

    client, err := anyi.NewClient("qwen-vision", config)
    if err != nil {
        log.Fatalf("创建客户端失败: %v", err)
    }

    // 准备包含图像的消息
    messages := []chat.Message{
        {
            Role:    "user",
            Content: "请详细描述这张图片中的内容。",
            Images: []string{
                "https://example.com/image.jpg", // 图像 URL
                // 或使用本地文件路径
                // "file:///path/to/local/image.jpg",
                // 或使用 base64 编码
                // "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAAAQ...",
            },
        },
    }

    // 发送请求
    response, info, err := client.Chat(messages, nil)
    if err != nil {
        log.Fatalf("聊天失败: %v", err)
    }

    fmt.Printf("图像描述: %s\n", response.Content)
    fmt.Printf("使用 tokens: %d\n", info.TotalTokens)
}
```

### 2. 图像分析工作流

```yaml
# multimodal_config.yaml
clients:
  - name: "qwen-vision"
    type: "dashscope"
    config:
      apiKey: "$DASHSCOPE_API_KEY"
      model: "qwen-vl-max"
      maxTokens: 1000

flows:
  - name: "image_analyzer"
    clientName: "qwen-vision"
    steps:
      - name: "describe_image"
        executor:
          type: "llm"
          withconfig:
            template: "详细描述这张图片的内容，包括物体、场景、颜色等。"
            systemMessage: "你是一个专业的图像分析师。"

      - name: "analyze_mood"
        executor:
          type: "llm"
          withconfig:
            template: "基于图片描述，分析图片传达的情感和氛围：{{.Text}}"
            systemMessage: "你是一个情感分析专家。"

      - name: "suggest_tags"
        executor:
          type: "llm"
          withconfig:
            template: "基于图片描述和情感分析，建议5-10个标签：\n描述：{{.Memory.description}}\n情感：{{.Text}}"
            systemMessage: "你是一个标签生成专家。"
```

加载并使用：

```go
package main

import (
    "fmt"
    "log"
    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/chat"
)

func main() {
    // 加载配置
    err := anyi.ConfigFromFile("multimodal_config.yaml")
    if err != nil {
        log.Fatal(err)
    }

    // 获取工作流
    flow, err := anyi.GetFlow("image_analyzer")
    if err != nil {
        log.Fatal(err)
    }

    // 创建包含图像的上下文
    context := anyi.NewFlowContext("")
    context.Images = []string{
        "https://example.com/beautiful-landscape.jpg",
    }

    // 运行工作流
    result, err := flow.Run(context)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("图像分析结果:\n%s\n", result.Text)
}
```

## 本地图像处理

### 1. 处理本地图像文件

```go
package main

import (
    "encoding/base64"
    "fmt"
    "io"
    "log"
    "os"
    "path/filepath"

    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/dashscope"
    "github.com/jieliu2000/anyi/llm/chat"
)

func imageToBase64(imagePath string) (string, error) {
    file, err := os.Open(imagePath)
    if err != nil {
        return "", err
    }
    defer file.Close()

    data, err := io.ReadAll(file)
    if err != nil {
        return "", err
    }

    // 根据文件扩展名确定 MIME 类型
    ext := filepath.Ext(imagePath)
    var mimeType string
    switch ext {
    case ".jpg", ".jpeg":
        mimeType = "image/jpeg"
    case ".png":
        mimeType = "image/png"
    case ".gif":
        mimeType = "image/gif"
    default:
        mimeType = "image/jpeg"
    }

    encoded := base64.StdEncoding.EncodeToString(data)
    return fmt.Sprintf("data:%s;base64,%s", mimeType, encoded), nil
}

func main() {
    config := dashscope.DefaultConfig(os.Getenv("DASHSCOPE_API_KEY"))
    config.Model = "qwen-vl-max"

    client, err := anyi.NewClient("qwen-vision", config)
    if err != nil {
        log.Fatal(err)
    }

    // 转换本地图像为 base64
    imageBase64, err := imageToBase64("./my-image.jpg")
    if err != nil {
        log.Fatalf("图像转换失败: %v", err)
    }

    messages := []chat.Message{
        {
            Role:    "user",
            Content: "这张图片中有什么？请详细描述。",
            Images:  []string{imageBase64},
        },
    }

    response, _, err := client.Chat(messages, nil)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("图像分析: %s\n", response.Content)
}
```

### 2. 批量图像处理

```go
package main

import (
    "fmt"
    "log"
    "os"
    "path/filepath"
    "sync"

    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/chat"
)

type ImageResult struct {
    Filename string
    Analysis string
    Error    error
}

func processImagesInDirectory(dir string, client anyi.Client) ([]ImageResult, error) {
    // 获取目录中的所有图像文件
    imageFiles, err := filepath.Glob(filepath.Join(dir, "*.{jpg,jpeg,png,gif}"))
    if err != nil {
        return nil, err
    }

    results := make([]ImageResult, len(imageFiles))
    var wg sync.WaitGroup

    // 并发处理图像
    for i, imageFile := range imageFiles {
        wg.Add(1)
        go func(index int, filename string) {
            defer wg.Done()

            result := ImageResult{Filename: filepath.Base(filename)}

            // 转换图像为 base64
            imageBase64, err := imageToBase64(filename)
            if err != nil {
                result.Error = err
                results[index] = result
                return
            }

            // 分析图像
            messages := []chat.Message{
                {
                    Role:    "user",
                    Content: "简要描述这张图片的主要内容（50字以内）。",
                    Images:  []string{imageBase64},
                },
            }

            response, _, err := client.Chat(messages, nil)
            if err != nil {
                result.Error = err
            } else {
                result.Analysis = response.Content
            }

            results[index] = result
        }(i, imageFile)
    }

    wg.Wait()
    return results, nil
}

func main() {
    // 设置客户端
    config := dashscope.DefaultConfig(os.Getenv("DASHSCOPE_API_KEY"))
    config.Model = "qwen-vl-max"

    client, err := anyi.NewClient("qwen-vision", config)
    if err != nil {
        log.Fatal(err)
    }

    // 处理目录中的所有图像
    results, err := processImagesInDirectory("./images", client)
    if err != nil {
        log.Fatal(err)
    }

    // 输出结果
    fmt.Println("图像分析结果:")
    for _, result := range results {
        if result.Error != nil {
            fmt.Printf("❌ %s: 错误 - %v\n", result.Filename, result.Error)
        } else {
            fmt.Printf("✅ %s: %s\n", result.Filename, result.Analysis)
        }
    }
}
```

## 使用 Ollama 进行本地多模态处理

### 1. 设置 Ollama 视觉模型

```bash
# 安装支持视觉的模型
ollama pull llava
ollama pull bakllava
ollama pull llava-phi3
```

### 2. 使用 Ollama 处理图像

```go
package main

import (
    "fmt"
    "log"

    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/ollama"
    "github.com/jieliu2000/anyi/llm/chat"
)

func main() {
    // 创建 Ollama 视觉客户端
    config := ollama.DefaultConfig("llava")
    config.BaseURL = "http://localhost:11434"

    client, err := anyi.NewClient("ollama-vision", config)
    if err != nil {
        log.Fatalf("创建 Ollama 客户端失败: %v", err)
    }

    // 转换本地图像
    imageBase64, err := imageToBase64("./test-image.jpg")
    if err != nil {
        log.Fatal(err)
    }

    messages := []chat.Message{
        {
            Role:    "user",
            Content: "请用中文描述这张图片中的内容。",
            Images:  []string{imageBase64},
        },
    }

    response, _, err := client.Chat(messages, nil)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("LLaVA 分析: %s\n", response.Content)
}
```

## 高级多模态应用

### 1. 图像问答系统

```go
package main

import (
    "bufio"
    "fmt"
    "log"
    "os"
    "strings"

    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/chat"
)

type ImageQASystem struct {
    client anyi.Client
    imageBase64 string
    conversation []chat.Message
}

func NewImageQASystem(client anyi.Client, imagePath string) (*ImageQASystem, error) {
    imageBase64, err := imageToBase64(imagePath)
    if err != nil {
        return nil, err
    }

    system := &ImageQASystem{
        client: client,
        imageBase64: imageBase64,
        conversation: []chat.Message{
            {
                Role:    "system",
                Content: "你是一个专业的图像分析助手。用户会向你展示一张图片并提问，请根据图片内容准确回答。",
            },
        },
    }

    return system, nil
}

func (qa *ImageQASystem) Ask(question string) (string, error) {
    // 添加用户问题（包含图像）
    userMessage := chat.Message{
        Role:    "user",
        Content: question,
        Images:  []string{qa.imageBase64},
    }
    qa.conversation = append(qa.conversation, userMessage)

    // 获取回答
    response, _, err := qa.client.Chat(qa.conversation, nil)
    if err != nil {
        return "", err
    }

    // 添加助手回答到对话历史
    assistantMessage := chat.Message{
        Role:    "assistant",
        Content: response.Content,
    }
    qa.conversation = append(qa.conversation, assistantMessage)

    return response.Content, nil
}

func main() {
    // 设置客户端
    config := dashscope.DefaultConfig(os.Getenv("DASHSCOPE_API_KEY"))
    config.Model = "qwen-vl-max"

    client, err := anyi.NewClient("qwen-vision", config)
    if err != nil {
        log.Fatal(err)
    }

    // 创建图像问答系统
    qa, err := NewImageQASystem(client, "./sample-image.jpg")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("图像问答系统已启动！输入 'quit' 退出。")
    fmt.Println("你可以问关于图片的任何问题。")

    scanner := bufio.NewScanner(os.Stdin)
    for {
        fmt.Print("\n问题: ")
        if !scanner.Scan() {
            break
        }

        question := strings.TrimSpace(scanner.Text())
        if question == "quit" {
            break
        }

        if question == "" {
            continue
        }

        answer, err := qa.Ask(question)
        if err != nil {
            fmt.Printf("错误: %v\n", err)
            continue
        }

        fmt.Printf("回答: %s\n", answer)
    }

    fmt.Println("再见！")
}
```

### 2. 图像比较分析

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/chat"
)

func compareImages(client anyi.Client, image1Path, image2Path string) (string, error) {
    // 转换两张图像为 base64
    image1Base64, err := imageToBase64(image1Path)
    if err != nil {
        return "", err
    }

    image2Base64, err := imageToBase64(image2Path)
    if err != nil {
        return "", err
    }

    messages := []chat.Message{
        {
            Role:    "user",
            Content: "请比较这两张图片的异同，包括内容、风格、颜色等方面的差异。",
            Images:  []string{image1Base64, image2Base64},
        },
    }

    response, _, err := client.Chat(messages, nil)
    if err != nil {
        return "", err
    }

    return response.Content, nil
}

func main() {
    config := dashscope.DefaultConfig(os.Getenv("DASHSCOPE_API_KEY"))
    config.Model = "qwen-vl-max"

    client, err := anyi.NewClient("qwen-vision", config)
    if err != nil {
        log.Fatal(err)
    }

    comparison, err := compareImages(client, "./image1.jpg", "./image2.jpg")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("图像比较结果:\n%s\n", comparison)
}
```

### 3. 图像内容提取工作流

```yaml
# image_extraction.yaml
clients:
  - name: "qwen-vision"
    type: "dashscope"
    config:
      apiKey: "$DASHSCOPE_API_KEY"
      model: "qwen-vl-max"

flows:
  - name: "extract_image_content"
    clientName: "qwen-vision"
    steps:
      - name: "extract_text"
        executor:
          type: "llm"
          withconfig:
            template: "提取图片中的所有文字内容，如果没有文字则返回'无文字'。"
            systemMessage: "你是一个OCR专家。"

      - name: "identify_objects"
        executor:
          type: "llm"
          withconfig:
            template: "列出图片中的主要物体和元素。"
            systemMessage: "你是一个物体识别专家。"

      - name: "analyze_scene"
        executor:
          type: "llm"
          withconfig:
            template: "描述图片的场景和环境。"
            systemMessage: "你是一个场景分析专家。"

      - name: "generate_summary"
        executor:
          type: "llm"
          withconfig:
            template: "基于以下信息生成图片的完整摘要：\n文字内容：{{.Memory.text}}\n物体列表：{{.Memory.objects}}\n场景描述：{{.Text}}"
            systemMessage: "你是一个内容整合专家。"
```

## 性能优化

### 1. 图像压缩

```go
package main

import (
    "bytes"
    "encoding/base64"
    "fmt"
    "image"
    "image/jpeg"
    "image/png"
    "os"
    "path/filepath"
    "strings"
)

func compressImage(imagePath string, quality int, maxWidth, maxHeight int) (string, error) {
    file, err := os.Open(imagePath)
    if err != nil {
        return "", err
    }
    defer file.Close()

    // 解码图像
    img, format, err := image.Decode(file)
    if err != nil {
        return "", err
    }

    // 调整图像大小
    if img.Bounds().Dx() > maxWidth || img.Bounds().Dy() > maxHeight {
        img = resizeImage(img, maxWidth, maxHeight)
    }

    // 压缩图像
    var buf bytes.Buffer
    switch strings.ToLower(format) {
    case "jpeg", "jpg":
        err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
    case "png":
        err = png.Encode(&buf, img)
    default:
        err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
    }

    if err != nil {
        return "", err
    }

    // 转换为 base64
    encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
    return fmt.Sprintf("data:image/jpeg;base64,%s", encoded), nil
}

func resizeImage(img image.Image, maxWidth, maxHeight int) image.Image {
    bounds := img.Bounds()
    width, height := bounds.Dx(), bounds.Dy()

    // 计算缩放比例
    scaleX := float64(maxWidth) / float64(width)
    scaleY := float64(maxHeight) / float64(height)
    scale := scaleX
    if scaleY < scaleX {
        scale = scaleY
    }

    if scale >= 1 {
        return img // 不需要缩放
    }

    newWidth := int(float64(width) * scale)
    newHeight := int(float64(height) * scale)

    // 简单的最近邻缩放（生产环境建议使用更好的算法）
    newImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
    for y := 0; y < newHeight; y++ {
        for x := 0; x < newWidth; x++ {
            srcX := int(float64(x) / scale)
            srcY := int(float64(y) / scale)
            newImg.Set(x, y, img.At(srcX, srcY))
        }
    }

    return newImg
}
```

### 2. 缓存图像分析结果

```go
package main

import (
    "crypto/md5"
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "time"

    "github.com/jieliu2000/anyi"
)

type CachedImageAnalyzer struct {
    client    anyi.Client
    cacheDir  string
    cacheTTL  time.Duration
}

type CacheEntry struct {
    Result    string    `json:"result"`
    Timestamp time.Time `json:"timestamp"`
}

func NewCachedImageAnalyzer(client anyi.Client, cacheDir string, cacheTTL time.Duration) *CachedImageAnalyzer {
    os.MkdirAll(cacheDir, 0755)
    return &CachedImageAnalyzer{
        client:   client,
        cacheDir: cacheDir,
        cacheTTL: cacheTTL,
    }
}

func (cia *CachedImageAnalyzer) AnalyzeImage(imagePath, prompt string) (string, error) {
    // 生成缓存键
    cacheKey := cia.generateCacheKey(imagePath, prompt)
    cachePath := filepath.Join(cia.cacheDir, cacheKey+".json")

    // 检查缓存
    if result, found := cia.getFromCache(cachePath); found {
        return result, nil
    }

    // 分析图像
    imageBase64, err := imageToBase64(imagePath)
    if err != nil {
        return "", err
    }

    messages := []chat.Message{
        {
            Role:    "user",
            Content: prompt,
            Images:  []string{imageBase64},
        },
    }

    response, _, err := cia.client.Chat(messages, nil)
    if err != nil {
        return "", err
    }

    // 保存到缓存
    cia.saveToCache(cachePath, response.Content)

    return response.Content, nil
}

func (cia *CachedImageAnalyzer) generateCacheKey(imagePath, prompt string) string {
    // 获取文件信息
    fileInfo, err := os.Stat(imagePath)
    if err != nil {
        return ""
    }

    // 生成唯一键
    data := fmt.Sprintf("%s-%d-%s", imagePath, fileInfo.ModTime().Unix(), prompt)
    hash := md5.Sum([]byte(data))
    return fmt.Sprintf("%x", hash)
}

func (cia *CachedImageAnalyzer) getFromCache(cachePath string) (string, bool) {
    data, err := os.ReadFile(cachePath)
    if err != nil {
        return "", false
    }

    var entry CacheEntry
    if err := json.Unmarshal(data, &entry); err != nil {
        return "", false
    }

    // 检查是否过期
    if time.Since(entry.Timestamp) > cia.cacheTTL {
        os.Remove(cachePath)
        return "", false
    }

    return entry.Result, true
}

func (cia *CachedImageAnalyzer) saveToCache(cachePath, result string) {
    entry := CacheEntry{
        Result:    result,
        Timestamp: time.Now(),
    }

    data, err := json.Marshal(entry)
    if err != nil {
        return
    }

    os.WriteFile(cachePath, data, 0644)
}
```

## 最佳实践

### 1. 图像处理建议

- **图像大小**：压缩大图像以减少 token 使用
- **图像质量**：保持足够的质量以确保分析准确性
- **批处理**：对多个图像使用并发处理
- **缓存**：缓存分析结果以避免重复请求

### 2. 成本优化

- **模型选择**：根据任务复杂度选择合适的模型
- **图像预处理**：在发送前进行适当的压缩和裁剪
- **提示优化**：使用清晰、具体的提示以获得更好的结果

### 3. 错误处理

```go
func robustImageAnalysis(client anyi.Client, imagePath, prompt string) (string, error) {
    maxRetries := 3
    var lastErr error

    for i := 0; i < maxRetries; i++ {
        result, err := analyzeImage(client, imagePath, prompt)
        if err == nil {
            return result, nil
        }

        lastErr = err
        log.Printf("分析失败 (尝试 %d/%d): %v", i+1, maxRetries, err)

        // 指数退避
        time.Sleep(time.Duration(1<<i) * time.Second)
    }

    return "", fmt.Errorf("重试 %d 次后仍然失败: %v", maxRetries, lastErr)
}
```

## 下一步

现在您已经掌握了多模态应用开发，可以：

1. 查看 [错误处理指南](../how-to/error-handling.md) 来构建更健壮的应用
2. 了解 [性能优化](../how-to/performance.md) 来提升应用性能
3. 探索 [安全最佳实践](../advanced/security.md) 来保护您的应用
4. 学习 [Web 集成](../how-to/web-integration.md) 来构建 Web 应用

通过多模态功能，您可以构建更丰富、更智能的 AI 应用程序！
