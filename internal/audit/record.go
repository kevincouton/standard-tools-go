package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// ErrNotFound is returned when no audit records exist in storage.
var ErrNotFound = errors.New("audit: no records found")

// DecisionRecord captures a single tool decision with hash-chaining metadata.
type DecisionRecord struct {
	RequestID      string    `json:"request_id"`
	RecordedAt     time.Time `json:"recorded_at"`
	ToolName       string    `json:"tool_name"`
	Input          any       `json:"input"`
	InputHash      string    `json:"input_hash"`
	Output         any       `json:"output"`
	OutputHash     string    `json:"output_hash"`
	Status         string    `json:"status"`
	Error          string    `json:"error,omitempty"`
	GitCommitSHA   string    `json:"git_commit_sha"`
	PackageVersion string    `json:"package_version"`
	RandomSeed     int64     `json:"random_seed"`
	PrevRecordHash string    `json:"prev_record_hash"`
	RecordHash     string    `json:"record_hash"`
}

func hashBytes(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

// hashAny returns a stable SHA-256 hash of any JSON-marshalable value.
func hashAny(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("marshal value for hashing: %w", err)
	}
	return hashBytes(b), nil
}

// HashRecord returns a stable SHA-256 hash of the record excluding its own RecordHash field.
func HashRecord(r DecisionRecord) (string, error) {
	r.RecordHash = ""
	b, err := json.Marshal(r)
	if err != nil {
		return "", fmt.Errorf("marshal record for hashing: %w", err)
	}
	return hashBytes(b), nil
}
