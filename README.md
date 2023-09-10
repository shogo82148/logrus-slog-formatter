# logrus-slog-formatter

logrus-slog-formatter is [log/slog](https://pkg.go.dev/log/slog) hook for [logrus](https://github.com/sirupsen/logrus).

## SYNOPSIS

```go
package main

import (
	"io"
	"log/slog"
	"os"

	sloghook "github.com/shogo82148/logrus-slog-hook"
	"github.com/sirupsen/logrus"
)

func main() {
	h := slog.NewTextHandler(os.Stderr, nil)
	logrus.AddHook(sloghook.New(h))

	// logrus-slog-hook outputs the logs into STDERR.
	// I recommend that disable the default output.
	logrus.SetFormatter(sloghook.NewFormatter())
	logrus.SetOutput(io.Discard)

	logrus.WithFields(logrus.Fields{
		"name": "joe",
		"age":  42,
	}).Error("Hello world!")

	// Output:
	// time=2023-09-10T23:32:41.229+09:00 level=ERROR msg="Hello world!" age=42 name=joe
}
```
