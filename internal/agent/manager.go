package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	aiagentmanager "easyHR/event/aiagentmanager"
	"easyHR/internal/agent/config"
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
	agentMsgProducer      *aiagentmanager.AgentMsgProducer
	mu                    sync.RWMutex // 读写锁，保护并发访问
	promptDir             string
}

// NewAiAgentManager 创建一个新的AiAgentManager实例
// 初始化内部映射和锁
func NewAiAgentManager(cfg config.AgentConfig, producer *aiagentmanager.AgentMsgProducer, l logger.LoggerV1) *AiAgentManager {
	a := &AiAgentManager{
		agentMsgProducer: producer,
		l:                l,
	}
	err := a.initAiAgentManager(cfg)
	if err != nil {
		panic(err)
	}
	return a
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
		// 读取提示词文件 (System Prompt 和 User Prompt)
		sysPrimaryReviewPrompt, err := os.ReadFile(filepath.Join(a.promptDir, "sysPrimaryReviewPrompt.xml"))
		if err != nil {
			panic(fmt.Sprintf("failed to read sysPrimaryReviewPrompt: %v", err))
		}
		sysSecondaryReviewPrompt, err := os.ReadFile(filepath.Join(a.promptDir, "sysSecondaryReviewPrompt.xml"))
		if err != nil {
			panic(fmt.Sprintf("failed to read sysSecondaryReviewPrompt: %v", err))
		}
		var ag *Agent
		switch agent.Role {
		case "primary":
			ag = newAgent(model, sid, agent.Role, string(sysPrimaryReviewPrompt))
			a.primaryReviewAgents[sid] = ag
		case "secondary":
			ag = newAgent(model, sid, agent.Role, string(sysSecondaryReviewPrompt))
			a.secondaryReviewAgents[sid] = ag
		}
	}
	return nil
}

// Analysis 调用Agents解析简历获取分析结果
// 接收简历文档path和邮件title，返回简历的分析结果和错误信息
func (a *AiAgentManager) Analysis(ctx context.Context, file string, title string) error {
	var responseMu sync.Mutex
	var SecReviewerMsgs []*Message
	var wg sync.WaitGroup

	a.l.Info("发送简历至SecondaryReviewer评审")
	for _, agent := range a.secondaryReviewAgents {
		agent := agent // Capture loop variable
		wg.Add(1)
		go func() {
			defer wg.Done()
			msg, err := agent.AddTask(ctx, file, SecondaryReviewPrompt)
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
	usrPrompt := a.ConstructPrimaryReviewerUsrPrompt(SecReviewerMsgs)
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
	for _, msg := range allMsgs {
		if msg == nil {
			continue
		}
		var modelType string
		// Try to find agent to get model type
		if ag, ok := a.secondaryReviewAgents[msg.SessionID]; ok {
			modelType = ag.model.GetModelType()
		} else if ag, ok := a.primaryReviewAgents[msg.SessionID]; ok {
			modelType = ag.model.GetModelType()
		}

		evt := aiagentmanager.ResponseReceivedEvent{
			SessionID: msg.SessionID,
			ModelType: modelType,
			Role:      msg.Role,
			Input:     msg.Input,
			Output:    msg.Content,
			CreatedAt: msg.CreatedAt,
		}
		if err := (*a.agentMsgProducer).AgentProduceResponseReceivedEvent(evt); err != nil {
			a.l.Error("failed to produce event",
				logger.Field{Key: "session_id", Val: msg.SessionID},
				logger.Field{Key: "error", Val: err},
			)
		}
	}

	return nil
}

// ConstructPrimaryReviewerUsrPrompt 构造PrimaryReviewer的User Prompt
// 接收SecondaryReviewer的分析结果，返回新的User Prompt

func (a *AiAgentManager) ConstructPrimaryReviewerUsrPrompt(msgs []*Message) string {
	var sb strings.Builder
	sb.WriteString(PrimaryReviewPrompt)
	sb.WriteString("\n\n以下是各次级评审员（Secondary Reviewer）的分析结果，请结合这些信息进行综合评审：\n\n")

	for _, msg := range msgs {
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
