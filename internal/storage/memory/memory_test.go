package memory_test

import (
	"context"
	"testing"

	"github.com/simonarnost/ariadne/internal/storage"
	"github.com/simonarnost/ariadne/internal/storage/memory"
	"github.com/simonarnost/ariadne/internal/tuple"
)

func TestReadWrite(t *testing.T) {
	tests := []struct {
		name      string
		seed      []string
		del       []string
		filter    storage.Filter
		wantCount int
	}{
		{
			name:      "match by object",
			seed:      []string{"document:readme#viewer@user:alice"},
			filter:    storage.Filter{ObjectType: "document", ObjectID: "readme"},
			wantCount: 1,
		},
		{
			name:      "filter matches nothing",
			seed:      []string{"document:readme#viewer@user:alice"},
			filter:    storage.Filter{ObjectType: "document", ObjectID: "other"},
			wantCount: 0,
		},
		{
			name:      "duplicate insert collapses to one",
			seed:      []string{"document:readme#viewer@user:alice", "document:readme#viewer@user:alice"},
			filter:    storage.Filter{ObjectType: "document"},
			wantCount: 1,
		},
		{
			name:      "delete removes the tuple",
			seed:      []string{"document:readme#viewer@user:alice"},
			del:       []string{"document:readme#viewer@user:alice"},
			filter:    storage.Filter{ObjectType: "document"},
			wantCount: 0,
		},
		{
			name: "empty filter matches all",
			seed: []string{
				"document:readme#viewer@user:alice",
				"document:budget#editor@user:bob",
			},
			filter:    storage.Filter{},
			wantCount: 2,
		},
		{
			name: "filter by subject finds across objects",
			seed: []string{
				"document:readme#viewer@user:alice",
				"document:budget#viewer@user:alice",
				"document:secret#viewer@user:bob",
			},
			filter:    storage.Filter{SubjectType: "user", SubjectID: "alice"},
			wantCount: 2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			s := memory.New()

			// Separate calls: within one Write, deletes run before inserts.
			if err := s.Write(ctx, parseAll(t, tc.seed), nil); err != nil {
				t.Fatalf("seed write: %v", err)
			}
			if err := s.Write(ctx, nil, parseAll(t, tc.del)); err != nil {
				t.Fatalf("delete write: %v", err)
			}

			got, err := s.Read(ctx, tc.filter)
			if err != nil {
				t.Fatalf("read: %v", err)
			}
			if len(got) != tc.wantCount {
				t.Errorf("Read returned %d tuples, want %d: %v", len(got), tc.wantCount, got)
			}
		})
	}
}

func parseAll(t *testing.T, strs []string) []tuple.Tuple {
	t.Helper()
	out := make([]tuple.Tuple, 0, len(strs))
	for _, s := range strs {
		tup, err := tuple.Parse(s)
		if err != nil {
			t.Fatalf("parse %q: %v", s, err)
		}
		out = append(out, tup)
	}
	return out
}
