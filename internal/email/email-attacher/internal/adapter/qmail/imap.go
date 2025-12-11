package qq

import (
	"easyHR/internal/email/email-attacher/internal/config"
	"easyHR/internal/email/email-attacher/internal/imap"

	"github.com/emersion/go-imap/v2/imapclient"
)

// connectIMAP 建立QQ邮箱IMAP连接
func connectIMAP(cfg config.IMAPConfig) (*imapclient.Client, error) {
	return imap.Connect(cfg.Addr, cfg.Username, cfg.Password)
}
