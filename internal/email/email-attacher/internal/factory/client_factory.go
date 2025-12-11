package factory

import (
	"fmt"
	"strings"

	"easyHR/internal/email/email-attacher/domain"
	qq "easyHR/internal/email/email-attacher/internal/adapter/qmail"
)

// NewEmailClient 根据服务商类型创建客户端实例
func NewEmailClient(providerType string) (domain.EmailClient, error) {
	switch strings.ToLower(providerType) {
	case "qq":
		return qq.NewQQEmailClient(), nil
	default:
		return nil, fmt.Errorf("不支持的服务商类型：%s", providerType)
	}
}
