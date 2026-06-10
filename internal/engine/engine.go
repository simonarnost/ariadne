package engine

import (
	"context"
	"fmt"

	"github.com/simonarnost/ariadne/internal/storage"
	"github.com/simonarnost/ariadne/internal/tuple"
)

type Engine struct {
	store storage.TupleStore
}

func New(store storage.TupleStore) *Engine {
	return &Engine{store: store}
}

func (e *Engine) Check(ctx context.Context, req CheckRequest) (CheckResponse, error) {
	filter := storage.Filter{
		ObjectType:      req.Object.Type,
		ObjectID:        req.Object.ID,
		Relation:        req.Relation,
		SubjectType:     req.Subject.Type,
		SubjectID:       req.Subject.ID,
		SubjectRelation: &req.Subject.Relation,
	}
	tuples, err := e.store.Read(ctx, filter)

	if err != nil {
		return CheckResponse{}, fmt.Errorf("cannot read from storage: %w", err)
	}

	return CheckResponse{Allowed: len(tuples) > 0}, nil
}

type CheckRequest struct {
	Object   tuple.Object
	Relation string
	Subject  tuple.Subject
}

type CheckResponse struct {
	Allowed bool
}
