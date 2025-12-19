package position

import "encoding/xml"

// JobPosition represents the structure of a job posting.
type JobPosition struct {
	XMLName          xml.Name        `xml:"JobPosition"`
	Title            string          `xml:"Title"`
	Department       string          `xml:"Department"`
	Description      string          `xml:"Description"`
	Responsibilities []string        `xml:"Responsibilities>Item"`
	Requirements     JobRequirements `xml:"Requirements"`
	Benefits         []string        `xml:"Benefits>Item"`
}

// JobRequirements details the specific requirements for the candidate.
type JobRequirements struct {
	Education  string   `xml:"Education"`
	Experience string   `xml:"Experience"`
	Skills     []string `xml:"Skills>Item"`
}
