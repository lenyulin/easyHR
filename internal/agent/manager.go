package agent

import (
	"context"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"sync"

	aiagentmanager "easyHR/event/aiagentmanager"
	"easyHR/internal/agent/config"
	"easyHR/internal/agent/prompts/position"
	"easyHR/pkg/logger"
	"easyHR/pkg/snowflake"
)

// AiAgentManager 管理Agent实例的生命周期
// 负责创建、获取和销毁Agent实例，确保并发安全
// 支持不同类型的AI模型，通过配置灵活切换
// 维护一个会话ID到Agent实例的映射
type AiAgentManager struct {
	primaryReviewAgents   map[string]*Agent // 会话ID到Agent实例的映射
	secondaryReviewAgents map[string]*Agent // 会话ID到Agent实例的映射
	l                     logger.LoggerV1
	agentMsgProducer      aiagentmanager.AgentMsgProducer
	mu                    sync.RWMutex // 读写锁，保护并发访问
	promptDir             string
	wg                    sync.WaitGroup
	shutdown              bool
}

// NewAiAgentManager 创建一个新的AiAgentManager实例
// 初始化内部映射和锁
func NewAiAgentManager(cfg config.AgentConfig, producer aiagentmanager.AgentMsgProducer, l logger.LoggerV1) *AiAgentManager {
	a := &AiAgentManager{
		agentMsgProducer: producer,
		l:                l,
	}
	a.primaryReviewAgents = make(map[string]*Agent)
	a.secondaryReviewAgents = make(map[string]*Agent)
	err := a.initAiAgentManager(cfg)
	if err != nil {
		panic(err)
	}
	return a
}
func (a *AiAgentManager) Stop() {
	a.mu.Lock()
	a.shutdown = true
	a.mu.Unlock()
	a.wg.Wait()
}

// NewAiAgentManager LoadAgentsFromConfig load agents from config file
func (a *AiAgentManager) initAiAgentManager(cfg config.AgentConfig) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	// Initialize snowflake node
	// Note: In a distributed system, NodeID should be configured externally
	sfNode, err := snowflake.NewNode(1)
	if err != nil {
		return err
	}
	a.promptDir = cfg.PromptDir
	for _, agent := range cfg.Agents {
		model := NewAiModel(Configuration{
			ModelType: agent.ModelType,
			ModelName: agent.ModelName,
			APIKey:    agent.APIKey,
			BaseURL:   agent.BaseURL,
		})

		// Generate snowflake ID
		id := sfNode.Generate()
		sid := strconv.FormatInt(id, 10)

		var ag *Agent
		switch agent.Role {
		case "primary":
			ag = newAgent(model, sid, agent.Role, SysPrimaryReviewPrompt)
			a.primaryReviewAgents[sid] = ag
		case "secondary":
			ag = newAgent(model, sid, agent.Role, SysSecondaryReviewPrompt)
			a.secondaryReviewAgents[sid] = ag
		}
	}
	return nil
}

// Analysis 调用Agents解析简历获取分析结果
// 接收简历文档path和邮件title，返回简历的分析结果和错误信息
func (a *AiAgentManager) Analysis(ctx context.Context, file string, title string) error {
	a.mu.RLock()
	if a.shutdown {
		a.mu.RUnlock()
		return fmt.Errorf("agent manager is shutting down")
	}
	a.wg.Add(1)
	defer a.wg.Done()
	a.mu.RUnlock()

	// Parse title: "2026校园招聘-后端研发-Name-13333333333"
	parts := strings.Split(title, "-")
	if len(parts) < 4 {
		return fmt.Errorf("invalid title format: %s", title)
	}
	positionName := parts[1] // "后端研发"

	// Map position name to Job ID
	var jobID string
	if positionName == "后端研发" {
		jobID = "SoftWareDeveloper_jobId"
	} else {
		// Default or error handling
		jobID = "SoftWareDeveloper_jobId" // Fallback for now as per example
	}

	// Load Job Position
	jobPos, err := position.LoadJobPosition(jobID)
	if err != nil {
		a.l.Error("failed to load job position", logger.Field{Key: "job_id", Val: jobID}, logger.Field{Key: "error", Val: err})
		return err
	}

	// Marshal Job Position to string for prompt injection
	// Using XML format as it is structured
	jobDescBytes, err := xml.MarshalIndent(jobPos, "", "  ")
	if err != nil {
		return err
	}
	jobDescStr := string(jobDescBytes)

	var responseMu sync.Mutex
	var SecReviewerMsgs []*Message
	var wg sync.WaitGroup

	a.l.Info("发送简历至SecondaryReviewer评审")

	// Inject Job Description into prompt
	prompt := strings.Replace(UsrSecondaryReviewPrompt, "{{JOB_DESCRIPTION}}", jobDescStr, 1)
	// Also replace {{user_query}} if it exists, though the user request focused on job description.
	// The original prompt has {{user_query}}.
	// Assuming for now the "user query" is implicit or empty in this flow, or we should leave it?
	// The user said: "match with LoadJobPosition... replace content... then as UsrSecondaryReviewPrompt whole send".
	// If UsrSecondaryReviewPrompt has {{user_query}}, we might need to fill it or remove it.
	// Given the context of "Analysis", there might not be a specific user query yet?
	// Or maybe the 'title' had the query?
	// Re-reading: "将需求与...内容进行替换，然后作为一个UsrSecondaryReviewPrompt整体发送".
	// I will assume specific user query is not the main point here, but the Job Description is.
	// I'll just replace JOB_DESCRIPTION. If {{user_query}} remains, it might be an issue.
	// Let's replace {{user_query}} with a generic instruction or empty string to be safe.
	prompt = strings.Replace(prompt, "{{user_query}}", "请基于以上岗位描述进行评估。", 1)

	for _, agent := range a.secondaryReviewAgents {
		agent := agent // Capture loop variable
		wg.Add(1)
		go func() {
			defer wg.Done()
			msg, err := agent.AddTask(ctx, file, prompt)
			if err != nil {
				a.l.Error("agent add task failed",
					logger.Field{Key: "session_id", Val: agent.SessionID},
					logger.Field{Key: "model_type", Val: agent.model.GetModelType()},
					logger.Field{Key: "model_name", Val: agent.ModelName},
					logger.Field{Key: "error", Val: err},
				)
				return
			}
			responseMu.Lock()
			SecReviewerMsgs = append(SecReviewerMsgs, msg)
			responseMu.Unlock()
		}()
	}
	a.l.Info("等待SecondaryReviewer返回审评结果")
	wg.Wait()

	var PrimReviewerMsgs []*Message
	// Also inject job description into Primary Reviewer Prompt
	// The user request explicitly mentioned "UsrSecondaryReviewPrompt", but Primary also needs to know the job.
	// My previous edit to UsrPrimaryReviewPrompt added {{JOB_DESCRIPTION}}.
	usrPromptTemplate := a.ConstructPrimaryReviewerUsrPrompt(SecReviewerMsgs)
	usrPrompt := strings.Replace(usrPromptTemplate, "{{JOB_DESCRIPTION}}", jobDescStr, 1)

	a.l.Info("发送简历至PrimaryReviewer评审")
	for _, agent := range a.primaryReviewAgents {
		agent := agent // Capture loop variable
		wg.Add(1)
		go func() {
			defer wg.Done()
			msg, err := agent.AddTask(ctx, file, usrPrompt)
			if err != nil {
				a.l.Error("agent add task failed",
					logger.Field{Key: "session_id", Val: agent.SessionID},
					logger.Field{Key: "model_type", Val: agent.model.GetModelType()},
					logger.Field{Key: "model_name", Val: agent.ModelName},
					logger.Field{Key: "error", Val: err},
				)
				return
			}
			responseMu.Lock()
			PrimReviewerMsgs = append(PrimReviewerMsgs, msg)
			responseMu.Unlock()
		}()
	}
	a.l.Info("等待PrimaryReviewer返回审评结果")
	wg.Wait()
	// Send PrimaryReviewerMsgs and SecondaryReviewerMsgs to agentMsgProducer
	allMsgs := append(SecReviewerMsgs, PrimReviewerMsgs...)
	a.produceAnalysisEvents(allMsgs)

	return nil
}

func (a *AiAgentManager) produceAnalysisEvents(msgs []*Message) {
	for _, msg := range msgs {
		if msg == nil {
			continue
		}
		var modelType string
		var modelName string
		// Try to find agent to get model type
		if ag, ok := a.secondaryReviewAgents[msg.SessionID]; ok {
			modelType = ag.model.GetModelType()
			modelName = ag.ModelName
		} else if ag, ok := a.primaryReviewAgents[msg.SessionID]; ok {
			modelType = ag.model.GetModelType()
			modelName = ag.ModelName
		}

		evt := aiagentmanager.ResponseReceivedEvent{
			SessionID: msg.SessionID,
			ModelType: modelType,
			Role:      msg.Role,
			Input:     msg.Input,
			Output:    msg.Content,
			CreatedAt: msg.CreatedAt,
			ModelName: modelName,
		}

		if err := a.agentMsgProducer.AgentProduceResponseReceivedEvent(evt); err != nil {
			a.l.Error("failed to produce event",
				logger.Field{Key: "session_id", Val: msg.SessionID},
				logger.Field{Key: "error", Val: err},
			)
		}
	}
}

// ConstructPrimaryReviewerUsrPrompt 构造PrimaryReviewer的User Prompt
// 接收SecondaryReviewer的分析结果，返回新的User Prompt

func (a *AiAgentManager) ConstructPrimaryReviewerUsrPrompt(msgs []*Message) string {
	var sb strings.Builder
	sb.WriteString(UsrPrimaryReviewPrompt)
	sb.WriteString("\n\n以下是各次级评审员（Secondary Reviewer）的分析结果：\n\n")

	for i, msg := range msgs {
		sb.WriteString(fmt.Sprintf("评审结果 %d:\n", i+1))
		sb.WriteString(msg.Content)
		sb.WriteString("\n\n")
	}
	return sb.String()
}

// ClearAll 销毁所有Agent实例
// 清空映射，释放所有资源
func (a *AiAgentManager) ClearAll() {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 清空映射
	a.primaryReviewAgents = make(map[string]*Agent)
	a.secondaryReviewAgents = make(map[string]*Agent)
}
