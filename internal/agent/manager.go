package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	aiagentmanager "easyHR/event/aiagentmanager"
	"easyHR/internal/agent/config"
	"easyHR/pkg/snowflake"
)

// AiAgentManager 管理Agent实例的生命周期
// 负责创建、获取和销毁Agent实例，确保并发安全
// 支持不同类型的AI模型，通过配置灵活切换
// 维护一个会话ID到Agent实例的映射
type AiAgentManager struct {
	primaryReviewAgents   map[string]*Agent // 会话ID到Agent实例的映射
	secondaryReviewAgents map[string]*Agent // 会话ID到Agent实例的映射
	agents                map[string]*Agent // 会话ID到Agent实例的映射
	producer              aiagentmanager.AgentMsgProducer
	mu                    sync.RWMutex // 读写锁，保护并发访问
	promptDir             string
}

// NewAiAgentManager 创建一个新的AiAgentManager实例
// 初始化内部映射和锁
func NewAiAgentManager(cfg config.AgentConfig, producer aiagentmanager.AgentMsgProducer) *AiAgentManager {
	a := &AiAgentManager{}
	err := a.initAiAgentManager(cfg, producer)
	if err != nil {
		panic(err)
	}
	return a
}

// NewAiAgentManager LoadAgentsFromConfig load agents from config file
func (a *AiAgentManager) initAiAgentManager(cfg config.AgentConfig, producer aiagentmanager.AgentMsgProducer) error {
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

	var SecReviewerMsgs []*Message
	var wg sync.WaitGroup
	wg.Add(len(a.secondaryReviewAgents))
	fmt.Println("发送简历至SecondaryReviewer评审")
	for _, agent := range a.secondaryReviewAgents {
		go func() {
			msg, err := agent.AddTask(ctx, file, SecondaryReviewPrompt)
			if err != nil {
				fmt.Println("agent add task\n session id:", agent.SessionID,
					" agent Model type:", agent.model.GetModelType(),
					"agent Model name:", agent.ModelName,
					"err detail:", err)
			}
			SecReviewerMsgs = append(SecReviewerMsgs, msg)
			defer wg.Done()
		}()
	}
	fmt.Println("等待SecondaryReviewer返回审评结果")
	wg.Wait()

	var PrimReviewerMsgs []*Message
	usrPrompt := a.ConstructSecondaryReviewerUsrPrompt(SecReviewerMsgs)
	fmt.Println("发送简历至PrimaryReviewer评审")
	wg.Add(len(a.primaryReviewAgents))
	for _, agent := range a.primaryReviewAgents {
		go func() {
			msg, err := agent.AddTask(ctx, file, usrPrompt)
			if err != nil {
				fmt.Println("agent add task\n session id:", agent.SessionID,
					" agent Model type:", agent.model.GetModelType(),
					"agent Model name:", agent.ModelName,
					"err detail:", err)
			}
			PrimReviewerMsgs = append(PrimReviewerMsgs, msg)
			defer wg.Done()
		}()
	}
	fmt.Println("等待PrimaryReviewer返回审评结果")
	wg.Wait()
	return nil
}

func (a *AiAgentManager) ConstructSecondaryReviewerUsrPrompt(msgs []*Message) string {
	//TODO Not finished
	panic("implement me")
	var res string
	res = fmt.Sprintf("%s", PrimaryReviewPrompt)
	for _, msg := range msgs {
		res = fmt.Sprintf("%s", msg.Content)
	}
	return res
}

// GetAgent 根据会话ID获取AIHelper实例
// 接收会话ID，返回对应的AIHelper实例和是否存在的标志
func (a *AiAgentManager) GetAgent(sessionID string) (*Agent, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	agent, exists := a.agents[sessionID]
	return agent, exists
}

// DestroyAgent 根据会话ID销毁Agent实例
// 接收会话ID，从映射中删除对应的Agent实例
func (a *AiAgentManager) DestroyAgent(sessionID string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 从映射中删除
	delete(a.agents, sessionID)
}

// ClearAll 销毁所有Agent实例
// 清空映射，释放所有资源
func (a *AiAgentManager) ClearAll() {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 清空映射
	a.agents = make(map[string]*Agent)
}

// GetSessionCount 获取当前管理器中的会话数量
// 返回映射的长度
func (a *AiAgentManager) GetSessionCount() int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return len(a.agents)
}
