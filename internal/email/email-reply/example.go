package emailreply

// Example 展示如何使用email-reply包
//
// 使用步骤：
// 1. 导入包：import "github.com/yourusername/easyHR/pkg/email-reply"
// 2. 初始化包：emailreply.Init(&Config{...})
// 3. 发送邮件：emailreply.SendReply(toEmail, &ReplyParams{...}, subject)
//
// 以下是一个完整的示例：
//
// func main() {
//     // 初始化配置
//     cfg := &emailreply.Config{
//         IMAPServer:   "imap.example.com",
//         IMAPPort:     993,
//         SMTPServer:   "smtp.example.com",
//         SMTPPort:     587,
//         Username:     "your-email@example.com",
//         Password:     "your-password",
//         UseTLS:       true,
//         FromEmail:    "your-email@example.com",
//         FromName:     "Your Company HR",
//     }
//
//     // 初始化包
//     if err := emailreply.Init(cfg); err != nil {
//         log.Fatalf("Failed to initialize email-reply package: %v", err)
//     }
//
//     // 准备邮件参数
//     params := &emailreply.ReplyParams{
//         Name:        "张三",
//         CompanyName: "示例公司",
//         Year:        "2024",
//         Semester:    "春季",
//         JobName:     "软件工程师",
//         Link:        "https://example.com/application-status",
//         HRName:      "李四",
//         HRPhone:     "13800138000",
//     }
//
//     // 发送邮件
//     toEmail := "zhangsan@example.com"
//     subject := "示例公司2024届春季校园招聘 - 简历收到确认"
//     if err := emailreply.SendReply(toEmail, params, subject); err != nil {
//         log.Fatalf("Failed to send email: %v", err)
//     }
//
//     fmt.Println("Email sent successfully!")
// }
