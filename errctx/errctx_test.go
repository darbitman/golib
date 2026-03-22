package errctx

import (
	"errors"
	"strings"
	"testing"
)

func TestErrorf(t *testing.T) {
	err := Errorf("something failed: %s", "reason")

	// Should contain the location prefix (this file + line number)
	if !strings.Contains(err.Error(), "errctx_test.go:") {
		t.Errorf("expected error to contain file location, got: %s", err.Error())
	}

	// Should contain the formatted message
	if !strings.Contains(err.Error(), "something failed: reason") {
		t.Errorf("expected error to contain formatted message, got: %s", err.Error())
	}

	// Should contain the separator
	if !strings.Contains(err.Error(), " > ") {
		t.Errorf("expected error to contain ' > ' separator, got: %s", err.Error())
	}
}

func TestErrorf_WrapsError(t *testing.T) {
	sentinel := errors.New("sentinel")
	err := Errorf("wrapped: %w", sentinel)

	// The wrapped error should be unwrappable
	if !errors.Is(err, sentinel) {
		t.Errorf("expected errors.Is to find sentinel, got: %s", err.Error())
	}
}

func TestErrorfSkip(t *testing.T) {
	// Call through a wrapper to test skip behavior
	err := wrapper("test message")

	// Should report this file (the caller of wrapper), not the wrapper itself
	if !strings.Contains(err.Error(), "errctx_test.go:") {
		t.Errorf("expected error to point to test file, got: %s", err.Error())
	}

	if !strings.Contains(err.Error(), "test message") {
		t.Errorf("expected error to contain message, got: %s", err.Error())
	}
}

// wrapper simulates a helper function that wraps ErrorfSkip
func wrapper(msg string) error {
	return ErrorfSkip(1, "wrapper: %s", msg)
}
