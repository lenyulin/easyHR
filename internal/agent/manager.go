package agent

import (
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
	agents   map[string]*Agent // 会话ID到Agent实例的映射
	producer aiagentmanager.AgentMsgProducer
	mu       sync.RWMutex // 读写锁，保护并发访问
}

// NewAiAgentManager 创建一个新的AiAgentManager实例
// 初始化内部映射和锁
func NewAiAgentManager() *AiAgentManager {
	return &AiAgentManager{
		agents: make(map[string]*Agent),
	}
}

// LoadAgentsFromConfig load agents from config file
func (a *AiAgentManager) NewAiAgentManager(agents []config.AgentDetail, producer aiagentmanager.AgentMsgProducer) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Initialize snowflake node
	// Note: In a distributed system, NodeID should be configured externally
	sfNode, err := snowflake.NewNode(1)
	if err != nil {
		return err
	}

	for _, agent := range agents {
		var model AIModel
		switch agent.ModelType {
		case "gpt4":
			model = NewGPT4Model(GPT4Config{
				APIKey:  agent.APIKey,
				BaseURL: agent.BaseURL,
			})
		default:
			model = NewGPT4Model(GPT4Config{
				APIKey:  agent.APIKey,
				BaseURL: agent.BaseURL,
			})
		}

		// Generate snowflake ID
		id := sfNode.Generate()
		sid := strconv.FormatInt(id, 10)

		var sysMsg string
		switch agent.Role {
		case "primary":
			sysMsg = PrimaryReviewPrompt
		case "secondary":
			sysMsg = SecondaryReviewPrompt
		}
		ag := NewAgent(model, sid, agent.Role, sysMsg)
		a.agents[sid] = ag
	}
	return nil
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
