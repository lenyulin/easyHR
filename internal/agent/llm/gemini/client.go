package gemini

import (
	"context"
	"easyHR/internal/agent/llm"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type gemini struct {
	client    *genai.Client
	modelName string
}

func (g *gemini) AnalyzeResume(ctx context.Context, role string, sysPrompt string, usrPrompt string, f string, opt ...interface{}) (*llm.ResumeAnalysis, error) {

	file, err := g.client.UploadFileFromPath(ctx, f, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer g.client.DeleteFile(ctx, file.Name)
	// 检查文件状态，确保处理完毕
	for {
		fileInfo, err := g.client.GetFile(ctx, file.Name)
		if err != nil {
			return nil, err
		}
		if fileInfo.State == genai.FileStateActive {
			break
		}
		if fileInfo.State == genai.FileStateFailed {
			return nil, errors.New("文件处理失败")
		}
		time.Sleep(1 * time.Second)
	}
	// 配置模型与 System Instruction
	model := g.client.GenerativeModel(g.modelName)

	// 将 sysPrimaryReviewPrompt.xml 的内容设置为 SystemInstruction
	model.SystemInstruction = genai.NewUserContent(genai.Text(sysPrompt))

	// 设置生成参数 (可选)
	model.SetTemperature(0.2) // 分析类任务建议较低的 temperature

	// 3. 配置模型参数
	model.ResponseMIMEType = "application/json"  // 强制 JSON
	model.ResponseSchema = NewEvaluationSchema() // 绑定 Schema

	// 调用 GenerateContent (混合 User Prompt 和 PDF)
	resp, err := model.GenerateContent(ctx,
		genai.FileData{URI: file.URI}, // 传入 PDF 的引用
		genai.Text(usrPrompt),         // 传入 User Prompt
	)
	if err != nil {
		return nil, errors.New("GenerateContentFailed err:" + err.Error())
	}

	// 解析结果
	var analysis llm.ResumeAnalysis
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				if txt, ok := part.(genai.Text); ok {
					fmt.Println("--- 原始 JSON 字符串 ---")
					fmt.Println(string(txt))

					// 6. Unmarshal 到 Go 结构体
					err := json.Unmarshal([]byte(txt), &analysis)
					if err != nil {
						fmt.Printf("JSON 解析失败: %v", err)
						continue
					}
				}
			}
		}
	}
	return &analysis, nil
}

func NewGeminiProvider(key string, modelName string) llm.LLMProvider {
	ctx := context.Background()
	// 1. 初始化客户端
	// 请将 "YOUR_API_KEY" 替换为你的实际 API Key
	client, err := genai.NewClient(ctx, option.WithAPIKey(key))
	if err != nil {
		log.Fatal(err)
	}
	return &gemini{
		client:    client,
		modelName: modelName,
	}
}
