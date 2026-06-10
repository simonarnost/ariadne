package engine_test

import (
	"context"
	"errors"
	"testing"

	"github.com/simonarnost/ariadne/internal/engine"
	"github.com/simonarnost/ariadne/internal/storage"
	"github.com/simonarnost/ariadne/internal/storage/memory"
	"github.com/simonarnost/ariadne/internal/tuple"
)

func TestCheckDirect(t *testing.T) {
	tests := []struct {
		name string
		seed []string
		ask  string // the object#relation@subject to Check
		want bool
	}{
		{
			name: "direct hit",
			seed: []string{"document:readme#viewer@user:alice"},
			ask:  "document:readme#viewer@user:alice",
			want: true,
		},
		{
			name: "empty store",
			ask:  "document:readme#viewer@user:alice",
			want: false,
		},
		{
			name: "wrong subject",
			seed: []string{"document:readme#viewer@user:alice"},
			ask:  "document:readme#viewer@user:bob",
			want: false,
		},
		{
			name: "wrong relation",
			seed: []string{"document:readme#viewer@user:alice"},
			ask:  "document:readme#editor@user:alice",
			want: false,
		},
		{
			name: "wrong object",
			seed: []string{"document:readme#viewer@user:alice"},
			ask:  "document:budget#viewer@user:alice",
			want: false,
		},
		{
			// A userset can itself be the direct subject; the pointer filter
			// must match its non-empty relation exactly.
			name: "userset subject matches",
			seed: []string{"document:readme#viewer@group:eng#member"},
			ask:  "document:readme#viewer@group:eng#member",
			want: true,
		},
		{
			name: "concrete subject != userset tuple",
			seed: []string{"document:readme#viewer@group:eng#member"},
			ask:  "document:readme#viewer@user:eng",
			want: false,
		},
		{
			// The path alice->group->readme exists, but a depth-1 check must
			// not follow it. Flips to true when M3 adds recursion.
			name: "userset not walked",
			seed: []string{"document:readme#viewer@group:eng#member", "group:eng#member@user:alice"},
			ask:  "document:readme#viewer@user:alice",
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			eng := engineWith(t, tc.seed...)

			got, err := eng.Check(context.Background(), request(t, tc.ask))
			if err != nil {
				t.Fatalf("Check: %v", err)
			}
			if got.Allowed != tc.want {
				t.Errorf("Check(%s) allowed=%v, want %v", tc.ask, got.Allowed, tc.want)
			}
		})
	}
}

func TestCheckStorageError(t *testing.T) {
	sentinel := errors.New("storage down")
	eng := engine.New(errStore{err: sentinel})

	_, err := eng.Check(context.Background(), request(t, "document:readme#viewer@user:alice"))
	if !errors.Is(err, sentinel) {
		t.Fatalf("Check error = %v, want wrapping %v", err, sentinel)
	}
}

// errStore fails every operation, to exercise Check's error path.
type errStore struct{ err error }

func (s errStore) Write(context.Context, []tuple.Tuple, []tuple.Tuple) error { return s.err }
func (s errStore) Read(context.Context, storage.Filter) ([]tuple.Tuple, error) {
	return nil, s.err
}

// engineWith returns an Engine backed by a store seeded with the given tuples.
func engineWith(t *testing.T, seed ...string) *engine.Engine {
	t.Helper()
	store := memory.New()
	if err := store.Write(context.Background(), parseAll(t, seed), nil); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return engine.New(store)
}

func request(t *testing.T, s string) engine.CheckRequest {
	t.Helper()
	tup, err := tuple.Parse(s)
	if err != nil {
		t.Fatalf("parse %q: %v", s, err)
	}
	return engine.CheckRequest{Object: tup.Object, Relation: tup.Relation, Subject: tup.Subject}
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
