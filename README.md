# caller
Go package caller provides a helper to make getting caller information simpler, and adds the ability to ignore functions or packages, a la `testing.Helper`.

[![Go Reference](https://pkg.go.dev/badge/github.com/gdey/caller.svg)](https://pkg.go.dev/github.com/gdey/caller)

Please look at the [example file](example_embed_test.go) to see how to make use of it.

```go

package main

import (
	"fmt"
	"strings"

	"github.com/gdey/caller"
)

type Log struct {
	caller.ACaller
}

func (l *Log) Init() {
	// Need to use the full name of TypeName.MethodName
	l.IgnoreFunction("Log.Info")
	l.IgnoreFunction("Log.log")
}

func (l Log) log(level, msg string) {
	frame := l.Caller()

	absFile := frame.File
	// We want to keep the filename independent of the root directory of the project.
	// so filter out anything before github.com
	idx := strings.Index(absFile, "github.com")
	if idx != -1 {
		absFile = absFile[idx:]
	}
	fmt.Printf("[%v]{%v:%v} %v\n", level, absFile, frame.Line, msg)
}

func FatalInfo(l Log, msg string) {
	// We use helper here as this function is not going to be called more then once.
	l.Helper()
	l.Info(msg)
	fmt.Println("Would exit here")
}

// DoubleMessageInfo gets called quite a bit, so, don't use l.Helper; and instead use
// l.IgnoreFunction()
func DoubleMessageInfo(l Log, msg string) {
	l.Info(msg)
	fmt.Println(msg)
}

// Info will log an info level message to the logger
func (l Log) Info(msg string) {
	l.log("INFO", msg)
}

func main() {
	var l Log
	l.Init()
	// Doubled message is called frequently, so have the modification to ignore list only
	// happen once. Note it's just the name of the function, and not TypeName.FunctionName
	// in this case.
	l.IgnoreFunction("DoubleMessageInfo")
	l.Info("First info message")
	DoubleMessageInfo(l, "This message is doubled")
	FatalInfo(l, "Last info message")
}
```

