// Package prdextract turns unstructured PRD text into candidate requirements
// for human review (REQ-MONO-067). It does not write the RTM.
package prdextract

import (
	"fmt"
	"strings"
)

// Format is the PRD input kind the extractor accepts.
type Format string

const (
	FormatText     Format = "text"
	FormatMarkdown Format = "markdown"
)

// Input is one PRD document for extraction (REQ-MONO-067aa).
type Input struct {
	Format  Format
	Content string
}

// Candidate is one extracted requirement statement with its source span
// (REQ-MONO-067ab). Emitting candidates does not write the RTM.
type Candidate struct {
	Statement string
	Start     int // byte offset into Content
	End       int // exclusive byte offset into Content
}

// Accept validates that the extractor can take this PRD input.
// Plain text and Markdown are both accepted.
func Accept(in Input) error {
	switch in.Format {
	case FormatText, FormatMarkdown:
		if strings.TrimSpace(in.Content) == "" {
			return fmt.Errorf("prd content is empty")
		}
		return nil
	default:
		return fmt.Errorf("unsupported prd format: %q", in.Format)
	}
}

// Extract emits candidate statements and source spans without writing an RTM.
func Extract(in Input) ([]Candidate, error) {
	if err := Accept(in); err != nil {
		return nil, err
	}
	content := in.Content
	var out []Candidate
	offset := 0
	for _, line := range strings.SplitAfter(content, "\n") {
		trimmed := strings.TrimSpace(line)
		startInLine := strings.Index(line, trimmed)
		if startInLine < 0 {
			offset += len(line)
			continue
		}
		absStart := offset + startInLine
		absEnd := absStart + len(trimmed)
		if looksLikeRequirement(trimmed) {
			out = append(out, Candidate{
				Statement: trimmed,
				Start:     absStart,
				End:       absEnd,
			})
		}
		offset += len(line)
	}
	return out, nil
}

func looksLikeRequirement(s string) bool {
	lower := strings.ToLower(strings.TrimLeft(s, "#*- "))
	return strings.HasPrefix(lower, "the system shall") ||
		strings.HasPrefix(lower, "shall ") ||
		strings.Contains(lower, " shall ")
}
