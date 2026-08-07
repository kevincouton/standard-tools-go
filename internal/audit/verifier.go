package audit

import (
	"context"
	"fmt"
)

// Verifier checks the integrity of the stored audit chain.
type Verifier struct {
	storage Storage
}

func NewVerifier(s Storage) *Verifier {
	return &Verifier{storage: s}
}

// VerifyChain verifies that the latest record's hash matches its contents.
// Currently this validates the tip of the chain; full chain traversal can be added later.
func (v *Verifier) VerifyChain(ctx context.Context) error {
	latest, err := v.storage.Latest(ctx)
	if err != nil {
		return err
	}
	if latest.RequestID == "" {
		return nil
	}
	expected, err := HashRecord(latest)
	if err != nil {
		return err
	}
	if expected != latest.RecordHash {
		return fmt.Errorf("latest record hash mismatch: expected %s, got %s", expected, latest.RecordHash)
	}
	return nil
}
