package gemini

import "github.com/google/generative-ai-go/genai"

func NewEvaluationSchema() *genai.Schema {
	return &genai.Schema{
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
		Required: []string{"candidate_name", "match_score", "skills", "projects", "summary"},
	}
}
