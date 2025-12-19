package agent

import (
	_ "embed"
)

// 系统默认提示词常量
// 这些提示词用于指导AI模型的行为和响应格式

// PrimaryReviewPrompt variables
//
//go:embed prompts/sysPrimaryReviewPrompt.xml
var SysPrimaryReviewPrompt string

//go:embed prompts/usrPrimaryReviewPrompt.xml
var UsrPrimaryReviewPrompt string

// SecondaryReviewPrompt variables
//
//go:embed prompts/sysSecondaryReviewPrompt.xml
var SysSecondaryReviewPrompt string

//go:embed prompts/usrSecondaryReviewPrompt.xml
var UsrSecondaryReviewPrompt string
