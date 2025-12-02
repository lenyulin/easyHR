package config

import (
	"easyHR/pkg/email-attacher/config"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// 支持的服务商类型
var supportedProviders = map[string]bool{
	"qq":      true,
	"netease": true,
	"gmail":   true,
	"outlook": true,
}

// 支持的日志级别
var supportedLogLevels = map[string]bool{
	"debug": true,
	"info":  true,
	"warn":  true,
	"error": true,
}

// ValidateConfig 校验外部配置合法性
func ValidateConfig(cfg *config.AppConfig) error {
	// 1. 附件保存路径校验
	if cfg.AttachmentSavePath == "" {
		return errors.New("附件保存路径（AttachmentSavePath）不能为空")
	}
	if err := checkPathWritable(cfg.AttachmentSavePath); err != nil {
		return fmt.Errorf("附件路径无效：%w", err)
	}

	// 2. 已处理邮件记录路径校验
	if cfg.ProcessedEmailsPath == "" {
		return errors.New("已处理邮件记录路径（ProcessedEmailsPath）不能为空")
	}
	processedDir := filepath.Dir(cfg.ProcessedEmailsPath)
	if err := checkPathWritable(processedDir); err != nil {
		return fmt.Errorf("已处理邮件路径无效：%w", err)
	}

	// 3. 轮询间隔校验
	if cfg.PollInterval <= 0 {
		return errors.New("轮询间隔（PollInterval）必须大于0秒")
	}

	// 4. 服务商列表校验
	if len(cfg.Providers) == 0 {
		return errors.New("服务商列表（Providers）不能为空")
	}
	for idx, prov := range cfg.Providers {
		provType := strings.ToLower(prov.Type)
		if provType == "" {
			return fmt.Errorf("第%d个服务商Type不能为空", idx+1)
		}
		if !supportedProviders[provType] {
			supportedStr := strings.Join(getMapKeys(supportedProviders), ", ")
			return fmt.Errorf("第%d个服务商类型不支持（支持：%s）", idx+1, supportedStr)
		}
		if prov.Config == nil || len(prov.Config) == 0 {
			return fmt.Errorf("第%d个服务商（%s）Config不能为空", idx+1, provType)
		}
		// 校验IMAP必填项
		if err := validateIMAPConfig(prov.Config, provType, idx+1); err != nil {
			return err
		}
	}

	// 5. 日志配置校验
	if cfg.LoggerConfig != nil {
		logLevel := strings.ToLower(cfg.LoggerConfig.Level)
		if !supportedLogLevels[logLevel] {
			levelsStr := strings.Join(getMapKeys(supportedLogLevels), ", ")
			return fmt.Errorf("日志级别不支持（支持：%s）", levelsStr)
		}
		if cfg.LoggerConfig.Path != "" {
			logDir := filepath.Dir(cfg.LoggerConfig.Path)
			if err := checkPathWritable(logDir); err != nil {
				return fmt.Errorf("日志路径无效：%w", err)
			}
		}
	}

	// 6. 重试配置校验
	if cfg.RetryConfig != nil {
		if cfg.RetryConfig.MaxAttempts < 1 {
			return errors.New("重试次数（MaxAttempts）必须≥1")
		}
		if cfg.RetryConfig.Interval < 0 {
			return errors.New("重试间隔（Interval）不能为负数")
		}
	}

	return nil
}

// 校验IMAP协议必填配置
func validateIMAPConfig(provConfig map[string]interface{}, provType string, idx int) error {
	requiredKeys := []string{"imap_addr", "port", "username", "password"}
	for _, key := range requiredKeys {
		val, exists := provConfig[key]
		if !exists {
			return fmt.Errorf("第%d个服务商（%s）缺少必填项：%s", idx, provType, key)
		}
		// 类型校验
		switch key {
		case "imap_addr", "username", "password":
			strVal, ok := val.(string)
			if !ok || strVal == "" {
				return fmt.Errorf("第%d个服务商（%s）的%s必须是非空字符串", idx, provType, key)
			}
		case "port":
			portVal, ok := val.(int)
			if !ok {
				return fmt.Errorf("第%d个服务商（%s）的port必须是整数（如993）", idx, provType)
			}
			if portVal < 1 || portVal > 65535 {
				return fmt.Errorf("第%d个服务商（%s）的port必须在1-65535之间", idx, provType)
			}
		}
	}
	return nil
}

// 检查路径是否可写（不存在则创建）
func checkPathWritable(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("创建目录失败：%w", err)
		}
	} else if err != nil {
		return fmt.Errorf("访问目录失败：%w", err)
	}

	// 测试写入权限
	tempFile, err := os.CreateTemp(path, ".test_writable")
	if err != nil {
		return fmt.Errorf("无写入权限：%w", err)
	}
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempFile.Name())
	}()
	return nil
}

// 获取map的key列表
func getMapKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
