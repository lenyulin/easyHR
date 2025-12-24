package agent

import (
	"bytes"
	"context"
	"easyHR/internal/agent/llm/gemini"
	"fmt"
	"io"
	"log"
	"net/http"
	"testing"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

var usrPrompt = "<instructions>\n请分析随附的简历文件，并将其与提供的【岗位描述】（Job Description）进行匹配度分析。\n你不需要输出任何Markdown标记，直接返回符合JSON Schema的纯JSON数据。\n\n结合以下规则进行分析：\n1. **Candidate Name**: 提取候选人全名。\n2. **Match Score**: 根据候选人技能与【岗位描述】中Requirements的匹配度给出一个0-100的评分。\n3. **Skills**: 列出提取到的硬技能。\n4. **Projects**: 详细分析项目经验，提取项目名称、描述、使用的技能，并给出你的专业评价。\n5. **Campus/Work Experience**: 分析校园经历和工作经历。\n6. **Summary**: 给出简短的综合评价理由。\n\n请确保分析深入具体，参考约束条件中的要求。\n</instructions>\n\n<job_description_placeholder>\n<JobPosition>\n  <Title>Software Developer</Title>\n  <Department>Engineering</Department>\n  <Description>&#xA;        We are seeking a skilled Software Developer to join our dynamic team. You will be responsible for designing and implementing high-quality software solutions.&#xA;    </Description>\n  <Responsibilities>\n    <Item>Design and develop scalable software applications.</Item>\n    <Item>Collaborate with cross-functional teams to define, design, and ship new features.</Item>\n    <Item>Write clean, maintainable, and efficient code.</Item>\n    <Item>Participate in code reviews and advocate for best practices.</Item>\n  </Responsibilities>\n  <Requirements>\n    <Education>Bachelor&#39;s degree in Computer Science or related field.</Education>\n    <Experience>3+ years of experience in software development.</Experience>\n    <Skills>\n      <Item>Proficiency in Go, Java, or Python.</Item>\n      <Item>Experience with distributed systems and microservices.</Item>\n      <Item>Knowledge of SQL and NoSQL databases.</Item>\n      <Item>Familiarity with containerization (Docker, Kubernetes).</Item>\n    </Skills>\n  </Requirements>\n  <Benefits>\n    <Item>Competitive salary and performance-based bonuses.</Item>\n    <Item>Comprehensive health, dental, and vision insurance.</Item>\n    <Item>Flexible work hours and remote work options.</Item>\n  </Benefits>\n</JobPosition>\n</job_description_placeholder>\n"
var sysPrompt = "<role>\n你是一个专业的架构师兼HR专家。你的任务是基于已有的简历分析结果或进一步的具体问题，提供更深入的见解。\n</role>\n\n<constraints>\n1. 延续主评审时的严格标准：专业、客观、不忽悠。\n2. 回答应具体针对用户的问题，不要泛泛而谈。\n3. 如果用户询问具体技术细节，请从架构师角度深度剖析该候选人的技术深度。\n4. 忽略简历中如“精通”、“熟悉”等主观描述，只关注实际的技能和项目经验。\n5. 必须逐个分析候选人的技能和经验点，严禁将多个不同维度的技能或经验合并分析。\n6. 忽略简历中如“大型”、“大规模”、“分布式”、“高并发”、“高性能”、“高可用”等无法具体量化的营销性词汇，除非有具体数据支持。\n7. 保持客观、中立，不带个人感情色彩。\n8. 保持格式清晰，重点突出。\n</constraints>\n"

// LoggingTransport 是一个自定义的 RoundTripper，用于拦截请求
type LoggingTransport struct {
	Transport http.RoundTripper
}

// RoundTrip 实现 http.RoundTripper 接口
func (t *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// 1. 只有是发送给 Gemini API 的 POST 请求才处理
	if req.Method == "POST" && req.Body != nil {
		// 读取 Body 的字节
		bodyBytes, _ := io.ReadAll(req.Body)

		// 打印请求的 JSON Payload
		// 注意：这会包含你的 Prompt 和 Configuration，但通常还没包含 API Key (Key 通常在 Query Param 或 Header 中)
		fmt.Printf("\n--- [DEBUG] Raw Request JSON ---\n%s\n--------------------------------\n", string(bodyBytes))

		// 重要：Body 是一个 ReadCloser，读完就没了。
		// 必须重新赋值一个新的 ReadCloser，否则后续 SDK 发送请求时会报错。
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// 2. 执行实际的 HTTP 请求
	return t.Transport.RoundTrip(req)
}
func TestGemini(t *testing.T) {
	ctx := context.Background()
	// 3. 创建一个使用我们自定义 Transport 的 HTTP Client
	httpClient := &http.Client{
		Transport: &LoggingTransport{
			// 使用默认的 Transport 作为底层传输
			Transport: http.DefaultTransport,
		},
	}
	client, err := genai.NewClient(ctx,
		option.WithAPIKey("AIzaSyC7858OJh5rfJ5hy5x3YZfP0uM0bRGUz9c"),
		option.WithHTTPClient(httpClient))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	file, err := client.UploadFileFromPath(ctx, "/Users/leiyulin/easyHR/easyHR/internal/agent/leiyulin_cv_b2.pdf", nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.DeleteFile(ctx, file.Name)
	// 配置模型与 System Instruction
	model := client.GenerativeModel("gemini-2.5-flash")

	// 将 sysPrimaryReviewPrompt.xml 的内容设置为 SystemInstruction
	model.SystemInstruction = genai.NewUserContent(genai.Text(sysPrompt))

	// 设置生成参数 (可选)
	model.SetTemperature(0.2) // 分析类任务建议较低的 temperature

	// 3. 配置模型参数
	model.ResponseMIMEType = "application/json"         // 强制 JSON
	model.ResponseSchema = gemini.NewEvaluationSchema() // 绑定 Schema

	// 调用 GenerateContent (混合 User Prompt 和 PDF)
	resp, err := model.GenerateContent(ctx,
		genai.FileData{URI: file.URI}, // 传入 PDF 的引用
		genai.Text(usrPrompt),         // 传入 User Prompt
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp)
}
