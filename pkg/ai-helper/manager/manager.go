package manager

import (
	"sync"

	"easyHR/pkg/ai-helper"
	"easyHR/pkg/ai-helper/models"
)

// AIHelperManager 管理AIHelper实例的生命周期
// 负责创建、获取和销毁AIHelper实例，确保并发安全
// 支持不同类型的AI模型，通过配置灵活切换
// 维护一个会话ID到AIHelper实例的映射

type AIHelperManager struct {
	helpers map[string]*aihelper.AIHelper // 会话ID到AIHelper实例的映射
	mu      sync.RWMutex                  // 读写锁，保护并发访问
}

// NewAIHelperManager 创建一个新的AIHelperManager实例
// 初始化内部映射和锁
func NewAIHelperManager() *AIHelperManager {
	return &AIHelperManager{
		helpers: make(map[string]*aihelper.AIHelper),
	}
}

// CreateAIHelper 创建一个新的AIHelper实例并保存到管理器中
// 接收模型类型、会话ID、API密钥和基础URL
// 根据模型类型创建相应的AI模型实例
// 目前支持gpt4模型，可扩展支持其他模型
func (m *AIHelperManager) CreateAIHelper(modelType, sessionID, apiKey, baseURL string) (*aihelper.AIHelper, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查会话ID是否已存在
	if _, exists := m.helpers[sessionID]; exists {
		// 会话已存在，返回现有实例
		return m.helpers[sessionID], nil
	}

	// 创建AI模型实例
	var model aihelper.AIModel
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

	// 创建AIHelper实例，使用默认的saveFunc
	helper := aihelper.NewAIHelper(model, sessionID, nil)

	// 保存到映射中
	m.helpers[sessionID] = helper

	return helper, nil
}

// GetAIHelper 根据会话ID获取AIHelper实例
// 接收会话ID，返回对应的AIHelper实例和是否存在的标志
func (m *AIHelperManager) GetAIHelper(sessionID string) (*aihelper.AIHelper, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	helper, exists := m.helpers[sessionID]
	return helper, exists
}

// DestroyAIHelper 根据会话ID销毁AIHelper实例
// 接收会话ID，从映射中删除对应的AIHelper实例
func (m *AIHelperManager) DestroyAIHelper(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 从映射中删除
	delete(m.helpers, sessionID)
}

// ClearAll 销毁所有AIHelper实例
// 清空映射，释放所有资源
func (m *AIHelperManager) ClearAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 清空映射
	m.helpers = make(map[string]*aihelper.AIHelper)
}

// GetSessionCount 获取当前管理器中的会话数量
// 返回映射的长度
func (m *AIHelperManager) GetSessionCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.helpers)
}