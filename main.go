package main

import (
	"context"
	cvhelper "easyHR/pkg/cv-helper"
	"easyHR/pkg/cv-helper/store"
	emailattacher "easyHR/pkg/email-attacher"
	"easyHR/pkg/email-attacher/config"
	emailattacherdomain "easyHR/pkg/email-attacher/domain"
	"easyHR/pkg/logger"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type CVHelperConfig struct {
	MongoURI   string `yaml:"mongo_uri"`
	Database   string `yaml:"database"`
	Collection string `yaml:"collection"`
}

type MainConfig struct {
	EmailAttacher config.AppConfig `yaml:"email_attacher"`
	CVHelper      CVHelperConfig   `yaml:"cv_helper"`
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

	// 初始化 Mongo Store
	mongoStore, err := store.NewMongoStore(mainCfg.CVHelper.MongoURI, mainCfg.CVHelper.Database, mainCfg.CVHelper.Collection)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	defer mongoStore.Close(ctx) // 注意：这里简单处理，实际应传入 context

	// 初始化下载器
	attacher, err := emailattacher.NewEmailAttacher(&mainCfg.EmailAttacher, log)
	if err != nil {
		panic(err)
	}

	// 流量控制
	var (
		mu       sync.Mutex
		isPaused bool
	)

	// 初始化 CV Helper
	cvHelper := cvhelper.NewCVHelper(mongoStore, log)
	cvChan := make(chan string, 100)

	cvHelper.Run(cvChan, func() {
		// 当 CVHelper 空闲时回调
		mu.Lock()
		defer mu.Unlock()
		if isPaused {
			log.Info("CVHelper 队列已空，恢复邮件轮询...")
			if err := attacher.Start(); err != nil {
				log.Error("恢复轮询失败: " + err.Error())
			} else {
				isPaused = false
			}
		}
	})

	// 注册回调
	attacher.OnAttachmentDownloaded(func(email emailattacherdomain.Email, att emailattacherdomain.Attachment, savePath string) {
		log.Info(fmt.Sprintf("附件下载成功: 邮件ID=%s,附件名=%s,路径=%s", strconv.Itoa(int(email.ID)), att.Name, savePath))

		// 检查通道是否已满（简单的背压控制）
		select {
		case cvChan <- savePath:
			// 成功发送
		default:
			// 通道已满，暂停轮询
			mu.Lock()
			if !isPaused {
				log.Info("CVHelper 队列已满，暂停邮件轮询...")
				attacher.Stop()
				isPaused = true
			}
			mu.Unlock()

			// 阻塞发送（确保不丢数据，但会阻塞当前 goroutine，即 poller 的 goroutine）
			// 注意：这里阻塞会导致 poller 暂停处理后续邮件，这正是我们想要的
			cvChan <- savePath
		}
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
	close(cvChan)
	log.Info("下载器已停止")
}
