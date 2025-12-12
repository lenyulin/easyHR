package llm

// ResumeAnalysis 是 LLM 分析简历后返回的标准结构
// 所有的实现（Gemini/OpenAI）都必须解析成这个结构返回
type ResumeAnalysis struct {
	CandidateName string   `json:"candidate_name"`
	MatchScore    int      `json:"match_score"`
	Skills        []string `json:"skills"`
	Summary       string   `json:"summary"`
}
