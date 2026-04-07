package errors

import (
	stdErrors "errors"
	"fmt"
	"testing"
)

type wrappedError struct {
	msg string
}

func (e *wrappedError) Error() string {
	return e.msg
}

func TestWrap(t *testing.T) {
	base := stdErrors.New("root")

	if got := Wrap(nil, "ctx"); got != nil {
		t.Fatalf("expected nil when wrapping nil, got %v", got)
	}

	err := Wrap(base, "context")
	if err == nil {
		t.Fatalf("expected error")
	}
	if got := err.Error(); got != "context: root" {
		t.Fatalf("unexpected wrapped error: %q", got)
	}
	if !Is(err, base) {
		t.Fatalf("expected wrapped error to match base")
	}
}

func TestWrapf(t *testing.T) {
	base := stdErrors.New("root")
	err := Wrapf(base, "query %s", "users")
	if err == nil {
		t.Fatalf("expected error")
	}
	if got := err.Error(); got != "query users: root" {
		t.Fatalf("unexpected wrapped error: %q", got)
	}
}

func TestNewAndNewf(t *testing.T) {
	if got := New("hello"); got.Error() != "hello" {
		t.Fatalf("unexpected New error: %v", got)
	}
	if got := Newf("hello %s", "world"); got.Error() != "hello world" {
		t.Fatalf("unexpected Newf error: %v", got)
	}
}

func TestAs(t *testing.T) {
	base := &wrappedError{msg: "boom"}
	err := fmt.Errorf("outer: %w", base)

	var target *wrappedError
	if !As(err, &target) {
		t.Fatalf("expected As to succeed")
	}
	if target != base {
		t.Fatalf("unexpected target: %#v", target)
	}
}
