package broker

import (
	"fmt"

	"github.com/rtmx-ai/rtmx/internal/database"
)

// RefuseCompleteFromResultsAlone enforces REQ-MONO-069e (rtmx 2.0.0): a results
// file alone is not authority to set COMPLETE.
func RefuseCompleteFromResultsAlone(fromResultsFile bool, newStatus database.Status) error {
	if fromResultsFile && newStatus == database.StatusComplete {
		return fmt.Errorf("verify: results file alone cannot set COMPLETE; submit a commit to the broker (`%s`)", BrokerInstallHint)
	}
	return nil
}
