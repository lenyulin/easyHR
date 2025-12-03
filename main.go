package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	emailattacher "easyHR/pkg/email-attacher"
	"easyHR/pkg/email-attacher/config"
	emailattacherdomain "easyHR/pkg/email-attacher/domain"
	"easyHR/pkg/logger"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type MainConfig struct {
	EmailAttacher config.AppConfig `yaml:"email_attacher"`
}

func InitLogger() logger.LoggerV1 {
	cfg := zap.NewDevelopmentConfig()
	err := viper.UnmarshalKey("log", &cfg)
	if err != nil {
		panic(err)
	}
	l, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	return logger.NewZapLogger(l)
}

func loadConfig() (*MainConfig, error) {
	data, err := os.ReadFile("./configs/mail_attacher_config.yaml")
	if err != nil {
		return nil, err
	}
	var cfg MainConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func main() {
	log := InitLogger()
	// 加载配置
	mainCfg, err := loadConfig()
	if err != nil {
		panic(err)
	}

	// 初始化下载器
	attacher, err := emailattacher.NewEmailAttacher(&mainCfg.EmailAttacher, log)
	if err != nil {
		panic(err)
	}

	// 注册回调
	attacher.OnAttachmentDownloaded(func(email emailattacherdomain.Email, att emailattacherdomain.Attachment, savePath string) {
		log.Info(fmt.Sprintf("附件下载成功: 邮件ID=%s,附件名=%s,路径=%s", strconv.Itoa(int(email.ID)), att.Name, savePath))
		// 可添加自定义逻辑：如上传到OSS、发送通知等
	})

	attacher.OnError(func(err error, provider string) {
		log.Error(fmt.Sprintf("服务商%s 出错：%v", provider, err))
	})

	// 启动下载器
	if err := attacher.Start(); err != nil {
		panic(err)
	}
	log.Info("下载器启动成功，开始轮询邮件...")

	// 监听退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// 优雅关闭
	attacher.Stop()
	log.Info("下载器已停止")
}
