# golib

`golib` is a collection of high-performance, idiomatic, and robust Go utilities emphasizing memory safety, asynchronous execution, and standard error handling patterns.

## Installation

```bash
go get github.com/darbitman/golib
```

## Packages

### 1. `async` (Asynchronous Tasks & Queues)

#### ⚡️ `Future[T]`
Safely submit background tasks natively without blocking. You can `.Await()` the generic response concurrently from multiple goroutines safely without duplicating work. Panics inside background tasks are intelligently recovered, packaged as standard `error` values, and gracefully propagated instead of crashing your application.

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/darbitman/golib/async"
)

func main() {
	// 1. Submit a background task cleanly
	future := async.Submit(context.Background(), func(ctx context.Context) (string, error) {
		time.Sleep(1 * time.Second)
		return "Hello from the Future!", nil
	})

	// 2. Do other main thread work...
	fmt.Println("Doing other things...")

	// 3. Block and await the async response predictably
	res, err := future.Await()
	if err != nil {
		panic(err)
	}

	fmt.Println(res) // "Hello from the Future!"
}
```

#### 📦 `BufferQueue[T]`
A high-throughput, thread-safe, slice-backed batching queue. Its `PopAll()` flush routine natively guarantees zero dynamic array-reallocation during massive traffic bursts and unblocks the Mutex lock in blazing-fast $O(1)$ time by offloading cleanup directly to the Go Garbage Collector.

```go
queue := async.NewBufferQueue[int]()

// Concurrently push massive amounts of traffic
go queue.Push(1)
go queue.Push(2)
go queue.Push(3)

// Flush the entire slice instantly O(1)
batch := queue.PopAll()
fmt.Println(batch) // [1, 2, 3] 

// The internal queue memory cleanly shrinks if traffic permanently stops, 
// natively preventing memory hoarding!
```

---

### 2. `errctx` (Source-Aware Error Formatting)
Provides zero-allocation error wrappers to trace bug origins immediately without repeatedly dropping into slow absolute filesystem pathing (`os.Getwd()`) on every error creation.

```go
package main

import (
	"fmt"
	"github.com/darbitman/golib/errctx"
)

func buildSocket() error {
	// Traces exactly where the error threw relative to your project root!
	return errctx.Errorf("failed negotiating tls handshake")
}

func main() {
	err := buildSocket()
	fmt.Println(err)
	// Output: main.go:10 > failed negotiating tls handshake
}
```

---

### 3. `todo` (Dynamic Technical Debt Logging)
Provides explicit codebase markers to highlight missing implementations or Edge-Case warnings dynamically into telemetry frameworks. Rather than burying incomplete code inside a `// TODO:` comment that no one ever reads, this strictly broadcasts technical debt into runtime monitoring! 

```go
package main

import (
	"github.com/darbitman/golib/todo"
)

func HandlePayment() {
	// Flags this function automatically to your engineering team!
	// Output: payments/handler.go:8 [HandlePayment] > TODO: implement
	todo.Implement()
}

func SwitchCase(val string) {
	switch val {
	case "pending": // Valid state
	default:
		// Emits specific custom warnings right to your stdout/logging system
		// Output: logic/switch.go:17 [SwitchCase] > TODO: handle unmapped edge case
		todo.Logf("handle unmapped edge case %s", val)
	}
}
```
