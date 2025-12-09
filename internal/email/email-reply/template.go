package emailreply

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// renderTemplate 渲染邮件模板，将参数替换为实际值
// 参数：
//
//	params - 包含模板所需动态数据的参数对象
//
// 返回值：
//
//	string - 渲染后的邮件内容
//	error - 渲染过程中遇到的错误，如果成功则返回nil
func renderTemplate(params *ReplyParams) (string, error) {
	// 获取当前文件所在目录
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// 构建模板文件的完整路径
	templatePath := filepath.Join(currentDir, "pkg", "email-reply", "template.md")

	// 检查模板文件是否存在
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		// 如果不存在，尝试另一种路径（相对于当前执行文件的路径）
		execPath, execErr := os.Executable()
		if execErr != nil {
			return "", fmt.Errorf("failed to get executable path: %w", execErr)
		}
		execDir := filepath.Dir(execPath)
		templatePath = filepath.Join(execDir, "pkg", "email-reply", "template.md")

		// 再次检查模板文件是否存在
		if _, err := os.Stat(templatePath); os.IsNotExist(err) {
			return "", fmt.Errorf("template file not found at %s", templatePath)
		}
	}

	// 读取模板文件
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template file: %w", err)
	}

	// 转换为字符串
	templateStr := string(templateContent)

	// 替换模板中的变量
	replaced := templateStr
	replaced = strings.ReplaceAll(replaced, "${name}", params.Name)
	replaced = strings.ReplaceAll(replaced, "${companyName}", params.CompanyName)
	replaced = strings.ReplaceAll(replaced, "${year}", params.Year)
	replaced = strings.ReplaceAll(replaced, "${semester}", params.Semester)
	replaced = strings.ReplaceAll(replaced, "${jobName}", params.JobName)
	replaced = strings.ReplaceAll(replaced, "${link}", params.Link)
	replaced = strings.ReplaceAll(replaced, "${hrName}", params.HRName)
	replaced = strings.ReplaceAll(replaced, "${hrPhone}", params.HRPhone)

	return replaced, nil
}
