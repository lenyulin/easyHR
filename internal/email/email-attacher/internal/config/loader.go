package config

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"easyHR/internal/email/email-attacher/config"
)

// InitInternalConfig 转换外部配置为内部配置
func InitInternalConfig(externalCfg *config.AppConfig) (*InternalConfig, error) {
	// 1. 构建轮询器配置
	pollerCfg, err := buildPollerConfig(externalCfg)
	if err != nil {
		return nil, fmt.Errorf("轮询器配置：%w", err)
	}

	// 2. 构建存储配置
	storageCfg, err := buildStorageConfig(externalCfg)
	if err != nil {
		return nil, fmt.Errorf("存储配置：%w", err)
	}
	retryCfg := buildRetryConfig(externalCfg.RetryConfig)

	// 3. 构建服务商配置
	providersCfg, err := buildProvidersConfig(externalCfg.Providers)
	if err != nil {
		return nil, fmt.Errorf("服务商配置：%w", err)
	}

	// 组装内部配置
	return &InternalConfig{
		PollerConfig:    *pollerCfg,
		StorageConfig:   *storageCfg,
		ProvidersConfig: providersCfg,
		RetryConfig:     retryCfg,
	}, nil
}

// 构建重试配置（补全默认值）
func buildRetryConfig(externalRetry *config.RetryConfig) RetryConfig {
	if externalRetry == nil {
		return RetryConfig{
			MaxAttempts: 3,
			Interval:    5 * time.Second,
		}
	}

	cfg := RetryConfig{
		MaxAttempts: externalRetry.MaxAttempts,
		Interval:    externalRetry.Interval,
	}
	if cfg.MaxAttempts < 1 {
		cfg.MaxAttempts = 3
	}
	if cfg.Interval < 0 {
		cfg.Interval = 5 * time.Second
	}
	return cfg
}

// 构建轮询器配置
func buildPollerConfig(externalCfg *config.AppConfig) (*PollerConfig, error) {
	return &PollerConfig{
		Interval: time.Duration(externalCfg.PollInterval) * time.Second,
	}, nil
}

// 构建存储配置（标准化为绝对路径）
func buildStorageConfig(externalCfg *config.AppConfig) (*StorageConfig, error) {
	attachPath, err := filepath.Abs(externalCfg.AttachmentSavePath)
	if err != nil {
		return nil, err
	}

	processedPath, err := filepath.Abs(externalCfg.ProcessedEmailsPath)
	if err != nil {
		return nil, err
	}

	return &StorageConfig{
		AttachmentSavePath:  attachPath,
		ProcessedEmailsPath: processedPath,
	}, nil
}

// 构建服务商配置（map转强类型）
func buildProvidersConfig(externalProviders []config.ProviderConfig) ([]ProviderConfig, error) {
	var internalProviders []ProviderConfig

	for _, externalProv := range externalProviders {
		provType := strings.ToLower(externalProv.Type)
		extCfg := externalProv.Config

		// 提取IMAP配置（已校验过类型，可安全断言）
		imapAddr := extCfg["imap_addr"].(string)
		port := extCfg["port"].(int)
		username := extCfg["username"].(string)
		password := extCfg["password"].(string)

		// 组装完整IMAP地址
		fullAddr := fmt.Sprintf("%s:%d", imapAddr, port)

		internalProviders = append(internalProviders, ProviderConfig{
			Type: provType,
			IMAPConfig: IMAPConfig{
				Addr:     fullAddr,
				Username: username,
				Password: password,
			},
		})
	}

	return internalProviders, nil
}
