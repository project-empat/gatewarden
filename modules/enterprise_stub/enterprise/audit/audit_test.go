package audit

import (
	"context"
	"errors"
	"testing"
)

func TestNoopProvider(t *testing.T) {
	p := NewProvider(nil)
	ctx := context.Background()

	// Reads degrade gracefully.
	events, err := p.Query(ctx, &Query{})
	if err != nil {
		t.Fatalf("Query error: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("Query = %v, want empty", events)
	}

	// Writes fail loudly.
	if err := p.Record(ctx, &Event{Action: "test"}); !errors.Is(err, ErrEnterpriseOnly) {
		t.Errorf("Record error = %v, want ErrEnterpriseOnly", err)
	}
	if _, err := p.Export(ctx, &Query{}, "json"); !errors.Is(err, ErrEnterpriseOnly) {
		t.Errorf("Export error = %v, want ErrEnterpriseOnly", err)
	}
	if _, err := p.Stream(ctx, "default"); !errors.Is(err, ErrEnterpriseOnly) {
		t.Errorf("Stream error = %v, want ErrEnterpriseOnly", err)
	}
}
