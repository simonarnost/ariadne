package memory

import (
	"context"
	"sync"

	"github.com/simonarnost/ariadne/internal/storage"
	"github.com/simonarnost/ariadne/internal/tuple"
)

type Store struct {
	mu     sync.RWMutex
	tuples map[tuple.Tuple]struct{}
}

var _ storage.TupleStore = (*Store)(nil)

func New() *Store {
	return &Store{tuples: make(map[tuple.Tuple]struct{})}
}

func (s *Store) Write(ctx context.Context, inserts, deletes []tuple.Tuple) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, t := range deletes {
		delete(s.tuples, t)
	}

	for _, t := range inserts {
		s.tuples[t] = struct{}{}
	}

	return nil
}

func (s *Store) Read(ctx context.Context, f storage.Filter) ([]tuple.Tuple, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var out []tuple.Tuple

	for t := range s.tuples {
		if f.Matches(t) {
			out = append(out, t)
		}
	}

	return out, nil
}
