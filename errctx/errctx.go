package errctx

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var rootDir string

func init() {
	rootDir, _ = os.Getwd()
}

// Errorf returns a formatted error annotated with the caller's
// source location (file:line). The file path is strictly made relative to the
// application's startup working directory when possible. The underlying error
// is wrapped with %w, so it remains compatible with [errors.Is] and [errors.As].
//
// If the caller information or working directory cannot be determined, the
// plain formatted error is returned without location context.
//
// Example output:
//
//	errctx/source.go:10 > an error occurred: something went wrong
func Errorf(format string, args ...any) error {
	return ErrorfSkip(1, format, args...)
}

// ErrorfSkip behaves like [Errorf] but allows the caller to control how
// many stack frames to skip when determining the source location. A skip
// of 0 reports the caller of ErrorfSkip itself, 1 reports that caller's
// caller, and so on.
//
// This is useful when Errorf is wrapped inside another helper function
// and the reported location should point to the original call site
// rather than the wrapper.
func ErrorfSkip(skip int, format string, args ...any) error {
	_, file, lineNo, ok := runtime.Caller(skip + 1)
	if !ok {
		return fmt.Errorf(format, args...)
	}

	path := file
	if rootDir != "" {
		if rel, err := filepath.Rel(rootDir, file); err == nil && !strings.HasPrefix(rel, "..") {
			path = rel
		}
	}

	return fmt.Errorf("%s:%d > "+format, append([]any{path, lineNo}, args...)...)
}
