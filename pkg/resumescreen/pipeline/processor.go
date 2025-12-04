package pipeline

import (
	"context"
	"easyHR/pkg/resumescreen/types"
	"io"
)

// Processor handles the ETL pipeline for resumes.
type Processor struct {
	ocr       OCRAdapter
	extractor Extractor
	cleaner   Cleaner
}

func NewProcessor(ocr OCRAdapter, extractor Extractor, cleaner Cleaner) *Processor {
	return &Processor{
		ocr:       ocr,
	extractor: extractor,
		cleaner:   cleaner,
	}
}

// Process takes a raw input (e.g., file stream) and returns a structured Candidate.
func (p *Processor) Process(ctx context.Context, input io.Reader, filename string) (*types.Candidate, error) {
	// 1. OCR
	text, err := p.ocr.ExtractText(ctx, input)
	if err != nil {
		return nil, err
	}

	// 2. Extract
	structured, err := p.extractor.Extract(ctx, text)
	if err != nil {
		return nil, err
	}

	// 3. Clean
	cleaned, err := p.cleaner.Clean(ctx, structured)
	if err != nil {
		return nil, err
	}

	// Construct Candidate
	candidate := &types.Candidate{
		// ID would be generated here or in storage
		Resume: types.ResumeData{
			RawText:    text,
			Structured: *cleaned,
		},
		// Other fields populated as needed
	}

	return candidate, nil
}
