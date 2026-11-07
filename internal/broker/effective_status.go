package broker

import (
	"crypto/ed25519"

	"github.com/rtmx-ai/rtmx/internal/attestation"
	"github.com/rtmx-ai/rtmx/internal/database"
)

// EffectiveStatus is the status readers must use (REQ-MONO-069f / 069g).
// A CSV COMPLETE cell counts only when the attestation signature verifies.
func EffectiveStatus(csvStatus database.Status, rec *attestation.Record, pub ed25519.PublicKey) database.Status {
	if csvStatus != database.StatusComplete {
		return csvStatus
	}
	if rec == nil || len(pub) != ed25519.PublicKeySize {
		return database.StatusPartial
	}
	if err := attestation.Verify(pub, *rec); err != nil {
		return database.StatusPartial
	}
	return database.StatusComplete
}
