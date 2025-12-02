package storage

import (
	"os"

	"easyHR/PKG/email-attacher/domain"
	"easyHR/pkg/logger"
)

// AttachmentStorage 附件存储（基于本地文件系统）
type AttachmentStorage struct {
	basePath string            // 附件存储根目录
	logger   *logger.ZapLogger // 日志器
}

// NewAttachmentStorage 创建附件存储实例
func NewAttachmentStorage(basePath string) *AttachmentStorage {
	return &AttachmentStorage{
		basePath: basePath,
		logger:   logger.MustNewDefaultLogger(),
	}
}

// GetBasePath 获取附件存储根目录
func (as *AttachmentStorage) GetBasePath() string {
	return as.basePath
}

// SaveAttachment 保存附件（调用客户端下载并写入本地）
func (as *AttachmentStorage) SaveAttachment(client domain.EmailClient, att domain.Attachment, savePath string) error {
	// 创建存储目录
	if err := os.MkdirAll(savePath, 0755); err != nil {
		return err
	}

	// 调用客户端下载附件
	if err := client.DownloadAttachment(att, savePath); err != nil {
		return err
	}

	as.logger.Info("附件保存成功", logger.Zap.String("name", att.Name), logger.Zap.String("path", savePath))
	return nil
}
