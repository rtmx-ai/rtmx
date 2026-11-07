package broker

import (
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/database"
)

func TestResultsFileAloneCannotSetComplete(t *testing.T) {
	err := RefuseCompleteFromResultsAlone(true, database.StatusComplete)
	if err == nil || !strings.Contains(err.Error(), "results file alone cannot set COMPLETE") {
		t.Fatalf("err = %v", err)
	}
	if err := RefuseCompleteFromResultsAlone(true, database.StatusPartial); err != nil {
		t.Fatalf("partial from results should be allowed for migration paths: %v", err)
	}
	if err := RefuseCompleteFromResultsAlone(false, database.StatusComplete); err != nil {
		t.Fatalf("broker-attested complete must be allowed: %v", err)
	}
}
