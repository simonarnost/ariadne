// Package tuple defines the core relation-tuple type.
package tuple

import (
	"fmt"
	"strings"
)

// Object is the resource a relation is about (document:readme).
type Object struct {
	Type string
	ID   string
}

func (o Object) String() string { return o.Type + ":" + o.ID }

// Subject holds a relation. An empty Relation is a concrete user (user:alice);
// a non-empty one is a userset (group:eng#member = the members of group:eng).
type Subject struct {
	Type     string
	ID       string
	Relation string
}

func (s Subject) IsUserset() bool { return s.Relation != "" }

func (s Subject) String() string {
	if s.IsUserset() {
		return s.Type + ":" + s.ID + "#" + s.Relation
	}
	return s.Type + ":" + s.ID
}

// Tuple is a single relationship: object#relation@subject.
type Tuple struct {
	Object   Object
	Relation string
	Subject  Subject
}

func (t Tuple) String() string {
	return t.Object.String() + "#" + t.Relation + "@" + t.Subject.String()
}

// Parse reads the string form object#relation@subject into a Tuple.
func Parse(s string) (Tuple, error) {
	objRel, subj, ok := strings.Cut(s, "@")
	if !ok {
		return Tuple{}, fmt.Errorf("tuple %q: missing '@subject'", s)
	}

	objStr, relation, ok := strings.Cut(objRel, "#")
	if !ok {
		return Tuple{}, fmt.Errorf("tuple %q: missing '#relation'", s)
	}

	obj, err := parseObject(objStr)
	if err != nil {
		return Tuple{}, fmt.Errorf("tuple %q: %w", s, err)
	}

	subject, err := parseSubject(subj)
	if err != nil {
		return Tuple{}, fmt.Errorf("tuple %q: %w", s, err)
	}

	if relation == "" {
		return Tuple{}, fmt.Errorf("tuple %q: empty relation", s)
	}

	return Tuple{Object: obj, Relation: relation, Subject: subject}, nil
}

func parseObject(s string) (Object, error) {
	typ, id, ok := strings.Cut(s, ":")
	if !ok {
		return Object{}, fmt.Errorf("object %q: missing ':id'", s)
	}
	if typ == "" || id == "" {
		return Object{}, fmt.Errorf("object %q: empty type or id", s)
	}
	return Object{Type: typ, ID: id}, nil
}

func parseSubject(s string) (Subject, error) {
	ref, relation, _ := strings.Cut(s, "#") // relation optional
	typ, id, ok := strings.Cut(ref, ":")
	if !ok {
		return Subject{}, fmt.Errorf("subject %q: missing ':id'", s)
	}
	if typ == "" || id == "" {
		return Subject{}, fmt.Errorf("subject %q: empty type or id", s)
	}
	return Subject{Type: typ, ID: id, Relation: relation}, nil
}
