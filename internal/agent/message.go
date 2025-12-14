package agent

import (
	"time"

	"github.com/google/generative-ai-go/genai"
)

var EvaluationSchema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"candidate_name": {
			Type:        genai.TypeString,
			Description: "The full name of the candidate found in the resume.",
		},
		"match_score": {
			Type:        genai.TypeInteger,
			Description: "A score from 0 to 100 indicating fit for the role.",
		},
		"skills": {
			Type: genai.TypeArray,
			Items: &genai.Schema{
				Type: genai.TypeString,
			},
			Description: "List of technical skills extracted.",
		},
		"projects": {
			Type: genai.TypeArray,
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"name": {
						Type:        genai.TypeString,
						Description: "The name of the project.",
					},
					"description": {
						Type:        genai.TypeString,
						Description: "A description of the project.",
					},
					"skills": {
						Type: genai.TypeArray,
						Items: &genai.Schema{
							Type: genai.TypeString,
						},
						Description: "List of technical skills used in the project.",
					},
					"comment": {
						Type:        genai.TypeString,
						Description: "Your comment on the project.",
					},
				},
				Required: []string{"name", "description", "skills", "comment"},
			},
		},
		"campus_experience": {
			Type: genai.TypeArray,
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"name": {
						Type:        genai.TypeString,
						Description: "The name of the campus experience.",
					},
					"description": {
						Type:        genai.TypeString,
						Description: "A description of the campus experience.",
					},
					"skills": {
						Type: genai.TypeArray,
						Items: &genai.Schema{
							Type: genai.TypeString,
						},
						Description: "List of skills used in the campus experience.",
					},
					"comment": {
						Type:        genai.TypeString,
						Description: "Your comment on the campus experience.",
					},
				},
				Required: []string{"name", "description", "skills", "comment"},
			},
		},
		"work_experience": {
			Type: genai.TypeArray,
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"name": {
						Type:        genai.TypeString,
						Description: "The name of the work experience/job/internal.",
					},
					"description": {
						Type:        genai.TypeString,
						Description: "A description of the work experience/job/internal.",
					},
					"skills": {
						Type: genai.TypeArray,
						Items: &genai.Schema{
							Type: genai.TypeString,
						},
						Description: "List of skills used in the work experience/job/internal.",
					},
					"comment": {
						Type:        genai.TypeString,
						Description: "Your comment on the work experience/job/internal.",
					},
				},
				Required: []string{"name", "description", "skills", "comment"},
			},
		},
		"summary": {
			Type:        genai.TypeString,
			Description: "Brief reasoning for the score.",
		},
	},
	// 定义必须存在的字段
	Required: []string{"candidate_name", "match_score", "skills", "projects", "campus_experience", "work_experience", "summary"},
}

type Project struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Skills      []string `json:"skills"`
	Comment     string   `json:"comment"`
}

type Experience struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Skills      []string `json:"skills"`
	Comment     string   `json:"comment"`
}

type CandidateEvaluation struct {
	CandidateName    string       `json:"candidate_name"`
	MatchScore       int          `json:"match_score"`
	Skills           []string     `json:"skills"`
	Projects         []Project    `json:"projects"`
	CampusExperience []Experience `json:"campus_experience"`
	WorkExperience   []Experience `json:"work_experience"`
	Summary          string       `json:"summary"`
}

// Message 定义消息持久化的结构
type Message struct {
	ID        uint   `gorm:"primaryKey;autoIncrement" json:"id" bson:"_id,omitempty"`
	SessionID string `gorm:"index;not null;type:varchar(36)" json:"session_id" bson:"session_id"`
	Role      string `gorm:"type:varchar(20)" json:"role" bson:"role"` // user, system, assistant
	Content   string `gorm:"type:text" json:"content" bson:"content"`

	// Structured evaluation result
	Evaluation *CandidateEvaluation `gorm:"serializer:json" json:"evaluation,omitempty" bson:"evaluation,omitempty"`

	Input     string    `gorm:"type:text" json:"input" bson:"input"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

type MessageProcessor struct{}

// NewMessageProcessor 创建一个新的MessageProcessor实例
// 初始化消息处理器，返回实例
func NewMessageProcessor() *MessageProcessor {
	return &MessageProcessor{}
}
