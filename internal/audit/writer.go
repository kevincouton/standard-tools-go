package audit

import "context"

// Writer records hash-chained decision records to the underlying storage.
type Writer struct {
	storage Storage
}

func NewWriter(storage Storage) *Writer {
	return &Writer{storage: storage}
}

// Write computes hashes and chains the record to the previous one before persisting it.
func (w *Writer) Write(ctx context.Context, r DecisionRecord) error {
	latest, err := w.storage.Latest(ctx)
	if err != nil {
		return err
	}
	if latest.RecordHash != "" {
		r.PrevRecordHash = latest.RecordHash
	}

	inputHash, err := hashAny(r.Input)
	if err != nil {
		return err
	}
	r.InputHash = inputHash

	outputHash, err := hashAny(r.Output)
	if err != nil {
		return err
	}
	r.OutputHash = outputHash

	recordHash, err := HashRecord(r)
	if err != nil {
		return err
	}
	r.RecordHash = recordHash

	return w.storage.Append(ctx, r)
}
