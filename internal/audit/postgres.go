package audit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresStorage persists decision records to PostgreSQL.
type PostgresStorage struct {
	pool *pgxpool.Pool
}

func NewPostgresStorage(pool *pgxpool.Pool) *PostgresStorage {
	return &PostgresStorage{pool: pool}
}

// Append stores a decision record. It expects hashes to be pre-computed by the caller.
func (p *PostgresStorage) Append(ctx context.Context, r DecisionRecord) error {
	raw := struct {
		Input  any `json:"input"`
		Output any `json:"output"`
	}{
		Input:  r.Input,
		Output: r.Output,
	}
	rawJSON, err := json.Marshal(raw)
	if err != nil {
		return fmt.Errorf("marshal raw: %w", err)
	}

	inputJSON, err := json.Marshal(r.Input)
	if err != nil {
		return fmt.Errorf("marshal input: %w", err)
	}
	outputJSON, err := json.Marshal(r.Output)
	if err != nil {
		return fmt.Errorf("marshal output: %w", err)
	}

	const query = `
		INSERT INTO audit_records (
			request_id, recorded_at, tool_name, input_hash, output_hash,
			status, error, git_commit_sha, package_version, random_seed,
			prev_record_hash, record_hash, input_json, output_json, raw
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err = p.pool.Exec(ctx, query,
		r.RequestID, r.RecordedAt, r.ToolName, r.InputHash, r.OutputHash,
		r.Status, nilable(r.Error), r.GitCommitSHA, r.PackageVersion, r.RandomSeed,
		r.PrevRecordHash, r.RecordHash, inputJSON, outputJSON, rawJSON,
	)
	if err != nil {
		return fmt.Errorf("insert audit record: %w", err)
	}
	return nil
}

// Latest returns the most recent record ordered by id DESC.
func (p *PostgresStorage) Latest(ctx context.Context) (DecisionRecord, error) {
	const query = `
		SELECT request_id, recorded_at, tool_name, input_hash, output_hash,
			status, error, git_commit_sha, package_version, random_seed,
			prev_record_hash, record_hash, input_json, output_json
		FROM audit_records
		ORDER BY id DESC
		LIMIT 1
	`
	return scanRecord(p.pool.QueryRow(ctx, query))
}

// GetByRequestID returns the record matching the given request_id.
func (p *PostgresStorage) GetByRequestID(ctx context.Context, requestID string) (DecisionRecord, error) {
	const query = `
		SELECT request_id, recorded_at, tool_name, input_hash, output_hash,
			status, error, git_commit_sha, package_version, random_seed,
			prev_record_hash, record_hash, input_json, output_json
		FROM audit_records
		WHERE request_id = $1
	`
	return scanRecord(p.pool.QueryRow(ctx, query, requestID))
}

func scanRecord(row pgx.Row) (DecisionRecord, error) {
	var r DecisionRecord
	var inputJSON, outputJSON []byte
	var errStr *string

	err := row.Scan(
		&r.RequestID, &r.RecordedAt, &r.ToolName, &r.InputHash, &r.OutputHash,
		&r.Status, &errStr, &r.GitCommitSHA, &r.PackageVersion, &r.RandomSeed,
		&r.PrevRecordHash, &r.RecordHash, &inputJSON, &outputJSON,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return DecisionRecord{}, ErrNotFound
	}
	if err != nil {
		return DecisionRecord{}, fmt.Errorf("scan audit record: %w", err)
	}

	if errStr != nil {
		r.Error = *errStr
	}

	if len(inputJSON) > 0 {
		r.Input = json.RawMessage(inputJSON)
	}
	if len(outputJSON) > 0 {
		r.Output = json.RawMessage(outputJSON)
	}

	return r, nil
}

func nilable(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
