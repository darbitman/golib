package todo

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// logFunc is the function used to emit TODO warnings. By default it writes
// to the default slog logger via [slog.Warn]. Call [SetLogFunc] to override.
var logFunc func(string) = func(msg string) {
	slog.Warn(msg)
}

var logFuncMu sync.RWMutex

// SetLogFunc sets the function used to emit TODO warnings. This allows
// consumers to wire in any logging framework (zerolog, slog, zap, etc.)
// without the library depending on one.
//
// SetLogFunc is thread-safe and can be called concurrently with logging.
// To prevent slow loggers from blocking other operations, the logging
// function itself is executed lock-free.
//
// Example with zerolog:
//
//	todo.SetLogFunc(func(msg string) {
//	    zerolog.Warn().Msg(msg)
//	})
func SetLogFunc(fn func(string)) {
	logFuncMu.Lock()
	defer logFuncMu.Unlock()
	logFunc = fn
}

func emit(msg string) {
	logFuncMu.RLock()
	fn := logFunc
	logFuncMu.RUnlock()
	fn(msg)
}

var (
	projectRootDir string
	rootOnce       sync.Once
)

func getProjectRoot() string {
	rootOnce.Do(func() {
		projectRootDir = findProjectRoot()
	})
	return projectRootDir
}

// Implement logs a warning indicating that a function needs to be
// implemented. It reports the caller's source location and function name.
//
// Example output:
//
//	pkg/server.go:42 [server.HandleRequest] > TODO: implement
func Implement() {
	emit(fmt.Sprintf("%s > TODO: implement", callerDetails(1)))
}

// Logf logs a formatted warning indicating a TODO with a custom message.
// It reports the caller's source location and function name.
//
// Example output:
//
//	pkg/server.go:42 [server.HandleRequest] > TODO: handle edge case
func Logf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	emit(fmt.Sprintf("%s > TODO: %s", callerDetails(1), msg))
}

// callerDetails retrieves the caller's relative file path, line number,
// and function name. skip is the number of additional frames to skip
// beyond callerDetails itself.
func callerDetails(skip int) string {
	pc, file, lineNumber, ok := runtime.Caller(skip + 1)
	if !ok {
		return "???:0 [???]"
	}

	funcParts := strings.Split(runtime.FuncForPC(pc).Name(), "/")
	funcName := funcParts[len(funcParts)-1]

	path := file
	root := getProjectRoot()
	if root != "" {
		if rel, err := filepath.Rel(root, file); err == nil && !strings.HasPrefix(rel, "..") {
			path = rel
		}
	}

	return fmt.Sprintf("%s:%d [%s]", path, lineNumber, funcName)
}

// findProjectRoot searches for the project root by looking for a go.mod file.
// Returns the root directory if found, or an empty string if not.
func findProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			return ""
		}
		dir = parentDir
	}
}
