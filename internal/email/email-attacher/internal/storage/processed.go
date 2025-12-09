package storage

import (
	"encoding/json"
	"os"
	"sync"

	"easyHR/pkg/logger"

	"github.com/emersion/go-imap/v2"
)

// ProcessedStorage 已处理邮件存储（基于文件）
type ProcessedStorage struct {
	filePath     string            // 存储文件路径
	processedIDs map[imap.UID]bool // 已处理邮件ID缓存
	mu           sync.RWMutex      // 读写锁（保证并发安全）
	logger       logger.LoggerV1
}

// NewProcessedStorage 创建已处理邮件存储实例
func NewProcessedStorage(filePath string, logger logger.LoggerV1) (*ProcessedStorage, error) {
	ps := &ProcessedStorage{
		filePath:     filePath,
		processedIDs: make(map[imap.UID]bool),
		logger:       logger,
	}

	// 加载已处理记录
	if err := ps.load(); err != nil {
		return nil, err
	}

	return ps, nil
}

// IsProcessed 检查邮件是否已处理
func (ps *ProcessedStorage) IsProcessed(emailID imap.UID) bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.processedIDs[emailID]
}

// MarkAsProcessed 标记邮件为已处理（并持久化）
func (ps *ProcessedStorage) MarkAsProcessed(emailID imap.UID) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.processedIDs[emailID] = true
	return ps.save()
}

// load 从文件加载已处理记录
func (ps *ProcessedStorage) load() error {
	// 文件不存在则创建空文件
	if _, err := os.Stat(ps.filePath); os.IsNotExist(err) {
		return ps.save() // 保存空map，创建文件
	}

	// 读取文件内容
	data, err := os.ReadFile(ps.filePath)
	if err != nil {
		return err
	}

	// 反序列化
	if err := json.Unmarshal(data, &ps.processedIDs); err != nil {
		return err
	}
	ps.logger.Info("已处理邮件记录加载成功",
		logger.Field{
			Key: "count",
			Val: len(ps.processedIDs),
		})
	return nil
}

// save 持久化已处理记录到文件
func (ps *ProcessedStorage) save() error {
	data, err := json.MarshalIndent(ps.processedIDs, "", "  ")
	if err != nil {
		return err
	}

	// 写入文件（覆盖）
	if err := os.WriteFile(ps.filePath, data, 0644); err != nil {
		return err
	}

	return nil
}
