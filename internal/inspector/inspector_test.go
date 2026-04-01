package inspector

import (
	"fmt"
	"testing"
	"time"
)

type mockFactory struct {
	typ string
}

func (m *mockFactory) Create(config ConnectionConfig) (Inspector, error) { return nil, nil }
func (m *mockFactory) GetType() string                                   { return m.typ }

func contains(slice []string, v string) bool {
	for _, s := range slice {
		if s == v {
			return true
		}
	}
	return false
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := NewRegistry()
	mf := &mockFactory{typ: "mock"}
	r.Register("mock", mf)

	got, ok := r.Get("mock")
	if !ok {
		t.Fatalf("expected factory to be found")
	}
	if got != mf {
		t.Fatalf("factory mismatch: got %v want %v", got, mf)
	}

	types := r.GetSupportedTypes()
	if !contains(types, "mock") {
		t.Fatalf("supported types missing mock: %v", types)
	}
}

func TestGlobalRegistry_RegisterAndGetFactory(t *testing.T) {
	typ := fmt.Sprintf("test-%d", time.Now().UnixNano())
	mf := &mockFactory{typ: typ}
	// register into global registry
	Register(typ, mf)

	got, ok := GetFactory(typ)
	if !ok {
		t.Fatalf("expected global factory to be found for %s", typ)
	}
	if got != mf {
		t.Fatalf("global factory mismatch: got %v want %v", got, mf)
	}

	types := GetSupportedTypes()
	if !contains(types, typ) {
		t.Fatalf("global supported types missing %s: %v", typ, types)
	}
}
