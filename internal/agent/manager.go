package agent

import (
	"easyHR/internal/ai-helper/models"
	"sync"
)

// AiAgentManager 管理AIHelper实例的生命周期
// 负责创建、获取和销毁AIHelper实例，确保并发安全
// 支持不同类型的AI模型，通过配置灵活切换
// 维护一个会话ID到AIHelper实例的映射

type AiAgentManager struct {
	agents map[string]*Agent // 会话ID到AIHelper实例的映射
	mu     sync.RWMutex      // 读写锁，保护并发访问
}

// NewAIHelperManager 创建一个新的AIHelperManager实例
// 初始化内部映射和锁
func NewAIHelperManager() *AiAgentManager {
	return &AiAgentManager{
		agents: make(map[string]*Agent),
	}
}

// CreateAgent 创建一个新的Agent实例并保存到管理器中
// 接收模型类型、会话ID、API密钥和基础URL
// 根据模型类型创建相应的AI模型实例
// 目前支持gpt4模型，可扩展支持其他模型
func (a *AiAgentManager) CreateAgent(modelType, sessionID, apiKey, baseURL string) (*Agent, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 检查会话ID是否已存在
	if _, exists := a.agents[sessionID]; exists {
		// 会话已存在，返回现有实例
		return a.agents[sessionID], nil
	}

	// 创建AI模型实例
	var model Model
	switch modelType {
	case "gpt4":
		// 创建GPT-4模型实例
		model = models.NewGPT4Model(models.GPT4Config{
			APIKey:  apiKey,
			BaseURL: baseURL,
		})
	default:
		// 目前只支持gpt4模型，可扩展
		// 这里可以添加更多模型类型的支持
		model = models.NewGPT4Model(models.GPT4Config{
			APIKey:  apiKey,
			BaseURL: baseURL,
		})
	}

	// 创建Agent实例，使用默认的saveFunc
	agent := NewAgent(model, sessionID, nil)

	// 保存到映射中
	a.agents[sessionID] = agent

	return agent, nil
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
