package proofbundle

import (
	"encoding/json"
	"fmt"
	"strings"
)

// RenderProjection serializes verification counts for a non-technical reader
// (REQ-MONO-065d). It does not include a Gantt chart, sprint board, or
// portfolio plan.
func RenderProjection(c Counts) (string, error) {
	raw, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encode projection: %w", err)
	}
	body := string(raw)
	for _, forbidden := range []string{"gantt", "sprint", "portfolio", "board", "chart"} {
		if strings.Contains(strings.ToLower(body), forbidden) {
			return "", fmt.Errorf("projection must not include %q", forbidden)
		}
	}
	return body + "\n", nil
}
