package doubao

type Response struct {
	ID                 string       `json:"id"`
	CreatedAt          float64      `json:"created_at"`
	Error              interface{}  `json:"error"`
	IncompleteDetails  interface{}  `json:"incomplete_details"`
	Instructions       interface{}  `json:"instructions"`
	Model              string       `json:"model"`
	Object             string       `json:"object"`
	Output             []OutputItem `json:"output"`
	ParallelToolCalls  interface{}  `json:"parallel_tool_calls"`
	Temperature        interface{}  `json:"temperature"`
	ToolChoice         interface{}  `json:"tool_choice"`
	Tools              interface{}  `json:"tools"`
	TopP               interface{}  `json:"top_p"`
	MaxOutputTokens    int          `json:"max_output_tokens"`
	PreviousResponseID interface{}  `json:"previous_response_id"`
	Thinking           interface{}  `json:"thinking"`
	ServiceTier        string       `json:"service_tier"`
	Status             string       `json:"status"`
	Text               interface{}  `json:"text"`
	Usage              Usage        `json:"usage"`
	Caching            Caching      `json:"caching"`
	Store              bool         `json:"store"`
	ExpireAt           int64        `json:"expire_at"`
}

type OutputItem struct {
	ID      string    `json:"id"`
	Summary []Summary `json:"summary,omitempty"`
	Content []Content `json:"content,omitempty"`
	Role    string    `json:"role,omitempty"`
	Type    string    `json:"type"`
	Status  string    `json:"status"`
}

type Summary struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

type Content struct {
	Text        string      `json:"text"`
	Type        string      `json:"type"`
	Annotations interface{} `json:"annotations"`
}

type Usage struct {
	InputTokens         int                 `json:"input_tokens"`
	InputTokensDetails  InputTokensDetails  `json:"input_tokens_details"`
	OutputTokens        int                 `json:"output_tokens"`
	OutputTokensDetails OutputTokensDetails `json:"output_tokens_details"`
	TotalTokens         int                 `json:"total_tokens"`
}

type InputTokensDetails struct {
	CachedTokens int `json:"cached_tokens"`
}

type OutputTokensDetails struct {
	ReasoningTokens int `json:"reasoning_tokens"`
}

type Caching struct {
	Type string `json:"type"`
}
