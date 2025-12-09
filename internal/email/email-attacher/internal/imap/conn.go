package imap

import (
	"crypto/tls"
	"strings"

	"github.com/emersion/go-imap/v2/imapclient"
)

// Connect 建立IMAP连接（通用逻辑，所有IMAP服务商复用）
func Connect(addr, username, password string) (*imapclient.Client, error) {
	// 连接IMAP服务器（启用TLS）
	options := &imapclient.Options{
		TLSConfig: &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         addr[:strings.Index(addr, ":")], // 提取主机名用于证书验证
		},
	}
	c, err := imapclient.DialTLS(addr, options)
	if err != nil {
		return nil, err
	}

	// 登录
	if err := c.Login(username, password).Wait(); err != nil {
		return nil, err
	}

	return c, nil
}

// Close 关闭IMAP连接
func Close(c *imapclient.Client) error {
	if c == nil {
		return nil
	}
	// 登出并关闭连接
	if err := c.Logout().Wait(); err != nil {
		return err
	}
	return c.Close()
}
