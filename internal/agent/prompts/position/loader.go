package position

import (
	"embed"
	"encoding/xml"
	"fmt"
)

//go:embed *.xml
var jobFiles embed.FS

// LoadJobPosition loads a job position from the embedded XML files by job ID.
// The jobID should be the filename without the .xml extension (e.g., "SoftWareDeveloper_jobId").
func LoadJobPosition(jobID string) (*JobPosition, error) {
	filename := jobID + ".xml"

	// Read the file from the embedded filesystem
	data, err := jobFiles.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read job file %s: %w", filename, err)
	}

	var jobPos JobPosition
	if err := xml.Unmarshal(data, &jobPos); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job position from %s: %w", filename, err)
	}

	return &jobPos, nil
}
