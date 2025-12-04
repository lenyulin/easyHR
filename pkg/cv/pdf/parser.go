package pdf

import (
	"bytes"
	"fmt"

	"github.com/ledongthuc/pdf"
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(filePath string) (string, error) {
	f, r, err := pdf.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open pdf: %w", err)
	}
	defer f.Close()

	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("failed to get plain text: %w", err)
	}

	_, err = buf.ReadFrom(b)
	if err != nil {
		return "", fmt.Errorf("failed to read text: %w", err)
	}

	return buf.String(), nil
}
