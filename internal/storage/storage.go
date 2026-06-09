// Package storage defines the TupleStore interface and shared filter.
package storage

import (
	"context"
	"errors"

	"github.com/simonarnost/ariadne/internal/tuple"
)

// ErrNotFound is returned when a single-tuple lookup finds nothing.
var ErrNotFound = errors.New("tuple not found")

// Filter selects tuples to read; a zero-valued field matches any value.
// SubjectRelation is a pointer because "" is meaningful (a concrete user), so
// nil is needed to mean "any".
type Filter struct {
	ObjectType      string
	ObjectID        string
	Relation        string
	SubjectType     string
	SubjectID       string
	SubjectRelation *string
}

// Matches reports whether t satisfies every constraint set on f.
func (f Filter) Matches(t tuple.Tuple) bool {
	switch {
	case f.ObjectType != "" && f.ObjectType != t.Object.Type:
		return false
	case f.ObjectID != "" && f.ObjectID != t.Object.ID:
		return false
	case f.Relation != "" && f.Relation != t.Relation:
		return false
	case f.SubjectType != "" && f.SubjectType != t.Subject.Type:
		return false
	case f.SubjectID != "" && f.SubjectID != t.Subject.ID:
		return false
	case f.SubjectRelation != nil && *f.SubjectRelation != t.Subject.Relation:
		return false
	}
	return true
}

type TupleStore interface {
	// Write applies deletes then inserts, atomically and idempotently.
	Write(ctx context.Context, inserts, deletes []tuple.Tuple) error
	Read(ctx context.Context, f Filter) ([]tuple.Tuple, error)
}
