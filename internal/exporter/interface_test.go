package exporter

import (
	"errors"
	"sort"
	"testing"

	"github.com/schema-export/schema-export/internal/model"
)

type testExporter struct{}

func (e *testExporter) Export([]model.Table, []model.View, []model.Procedure, []model.Function, []model.Trigger, []model.Sequence, ExportOptions) error {
	return nil
}

func (e *testExporter) GetName() string {
	return "test"
}

func (e *testExporter) GetExtension() string {
	return ".test"
}

type testExporterFactory struct {
	exporter Exporter
	err      error
}

func (f *testExporterFactory) Create() (Exporter, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.exporter != nil {
		return f.exporter, nil
	}
	return &testExporter{}, nil
}

func (f *testExporterFactory) GetType() string {
	return "test"
}

func TestRegistry(t *testing.T) {
	r := NewRegistry()
	factory := &testExporterFactory{}

	r.Register("alpha", factory)
	got, ok := r.Get("alpha")
	if !ok {
		t.Fatalf("expected factory to be registered")
	}
	if got != factory {
		t.Fatalf("unexpected factory pointer")
	}

	r.Register("beta", factory)
	types := r.GetSupportedTypes()
	sort.Strings(types)
	if len(types) != 2 || types[0] != "alpha" || types[1] != "beta" {
		t.Fatalf("unexpected supported types: %#v", types)
	}
}

func TestGlobalRegistryHelpers(t *testing.T) {
	factory := &testExporterFactory{}
	Register("global-test", factory)

	got, ok := GetFactory("global-test")
	if !ok {
		t.Fatalf("expected factory from global registry")
	}
	if got != factory {
		t.Fatalf("unexpected factory from global registry")
	}

	types := GetSupportedTypes()
	found := false
	for _, typ := range types {
		if typ == "global-test" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected global type in %#v", types)
	}
}

func TestFactoryCreateErrorPropagation(t *testing.T) {
	expected := errors.New("boom")
	factory := &testExporterFactory{err: expected}
	got, err := factory.Create()
	if err != expected {
		t.Fatalf("expected create error to propagate, got %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil exporter on error")
	}
}
