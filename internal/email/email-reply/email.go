package emailreply

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
)

// sendEmail 发送邮件
// 参数：
//
//	to - 收件人邮箱地址
//	subject - 邮件主题
//	body - 邮件内容
//
// 返回值：
//
//	error - 发送过程中遇到的错误，如果成功则返回nil
func sendEmail(to, subject, body string) error {
	// 配置信息
	cfg := config

	// 构建邮件头部
	msg := fmt.Sprintf("From: %s <%s>\r\n", cfg.FromName, cfg.FromEmail)
	msg += fmt.Sprintf("To: %s\r\n", to)
	msg += fmt.Sprintf("Subject: %s\r\n", subject)
	msg += "Content-Type: text/plain; charset=utf-8\r\n"
	msg += "\r\n"
	msg += body

	// SMTP服务器地址
	smtpAddr := fmt.Sprintf("%s:%d", cfg.SMTPServer, cfg.SMTPPort)

	// 认证信息
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.SMTPServer)

	// 发送邮件
	if cfg.UseTLS {
		// 使用TLS连接
		tlsConfig := &tls.Config{
			ServerName:         cfg.SMTPServer,
			InsecureSkipVerify: false, // 不跳过证书验证
		}

		// 建立TLS连接
		conn, dialErr := tls.Dial("tcp", smtpAddr, tlsConfig)
		if dialErr != nil {
			return fmt.Errorf("failed to dial SMTP server with TLS: %w", dialErr)
		}
		defer conn.Close()

		// 创建SMTP客户端
		client, newErr := smtp.NewClient(conn, cfg.SMTPServer)
		if newErr != nil {
			return fmt.Errorf("failed to create SMTP client: %w", newErr)
		}
		defer client.Close()

		// 认证
		if authErr := client.Auth(auth); authErr != nil {
			return fmt.Errorf("failed to authenticate with SMTP server: %w", authErr)
		}

		// 设置发件人
		if fromErr := client.Mail(cfg.FromEmail); fromErr != nil {
			return fmt.Errorf("failed to set from address: %w", fromErr)
		}

		// 设置收件人
		if toErr := client.Rcpt(to); toErr != nil {
			return fmt.Errorf("failed to set to address: %w", toErr)
		}

		// 发送数据
		writer, dataErr := client.Data()
		if dataErr != nil {
			return fmt.Errorf("failed to get data writer: %w", dataErr)
		}

		_, writeErr := writer.Write([]byte(msg))
		if writeErr != nil {
			return fmt.Errorf("failed to write message: %w", writeErr)
		}

		if closeErr := writer.Close(); closeErr != nil {
			return fmt.Errorf("failed to close data writer: %w", closeErr)
		}

		// 发送QUIT命令
		if quitErr := client.Quit(); quitErr != nil {
			return fmt.Errorf("failed to quit SMTP client: %w", quitErr)
		}
	} else {
		// 不使用TLS连接
		if sendErr := smtp.SendMail(smtpAddr, auth, cfg.FromEmail, []string{to}, []byte(msg)); sendErr != nil {
			return fmt.Errorf("failed to send email via SMTP: %w", sendErr)
		}
	}

	return nil
}
