package pipeline

import (
	"context"
	"io"
)

// OCRAdapter defines the interface for OCR services.
type OCRAdapter interface {
	ExtractText(ctx context.Context, reader io.Reader) (string, error)
}
