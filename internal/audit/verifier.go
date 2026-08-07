package audit

import (
	"context"
	"errors"
	"fmt"
)

// Verifier checks the integrity of the stored audit chain.
type Verifier struct {
	storage Storage
}

func NewVerifier(s Storage) *Verifier {
	return &Verifier{storage: s}
}

// VerifyChain verifies the tip of the audit chain by recomputing the input,
// output, and record hashes and comparing them to the stored values.
// Full chain traversal can be added later.
func (v *Verifier) VerifyChain(ctx context.Context) error {
	latest, err := v.storage.Latest(ctx)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil
		}
		return fmt.Errorf("fetch latest audit record: %w", err)
	}

	if expected, err := hashAny(latest.Input); err != nil {
		return fmt.Errorf("hash stored input: %w", err)
	} else if expected != latest.InputHash {
		return fmt.Errorf("input hash mismatch: expected %s, got %s", expected, latest.InputHash)
	}

	if expected, err := hashAny(latest.Output); err != nil {
		return fmt.Errorf("hash stored output: %w", err)
	} else if expected != latest.OutputHash {
		return fmt.Errorf("output hash mismatch: expected %s, got %s", expected, latest.OutputHash)
	}

	expected, err := HashRecord(latest)
	if err != nil {
		return fmt.Errorf("hash stored record: %w", err)
	}
	if expected != latest.RecordHash {
		return fmt.Errorf("record hash mismatch: expected %s, got %s", expected, latest.RecordHash)
	}
	return nil
}
