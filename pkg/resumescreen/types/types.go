package types

import "time"

// Candidate represents a job candidate.
type Candidate struct {
	ID        string
	Name      string
	Email     string
	Phone     string
	Resume    ResumeData
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ResumeData contains the raw and structured data of a resume.
type ResumeData struct {
	RawText    string
	Structured StructuredResume
}

// StructuredResume represents the parsed structure of a resume.
type StructuredResume struct {
	Education  []Education
	Experience []Experience
	Skills     []string
	Projects   []Project
}

// Education represents an educational background.
type Education struct {
	School    string
	Degree    string
	Major     string
	StartDate time.Time
	EndDate   time.Time
}

// Experience represents a work experience.
type Experience struct {
	Company     string
	Title       string
	Description string
	StartDate   time.Time
	EndDate     time.Time
}

// Project represents a project experience.
type Project struct {
	Name        string
	Description string
	Role        string
}

// SearchResult represents a single result from a resume search.
type SearchResult struct {
	Candidate   Candidate
	Score       float64
	MatchReason string
}
