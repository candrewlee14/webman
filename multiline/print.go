package multiline

import (
	"fmt"
	"io"
	"sync"
)

const esc = "\033["

var (
	clearLine = []byte(esc + "2K\r")
	moveUp    = []byte(esc + "1F")
	moveDown  = []byte(esc + "1E")
)

type LineLogger struct {
	index int
	count int
	w     io.Writer
}

func NewLineLogger(index int, count int, w io.Writer) LineLogger {
	return LineLogger{index, count, w}
}

func (l *LineLogger) Printf(format string, a ...any) {
	for i := 0; i < l.count-l.index; i++ {
		fmt.Fprintf(l.w, "%s", moveUp)
	}
	fmt.Fprintf(l.w, "%s", clearLine)
	fmt.Fprintf(l.w, format, a...)
	for i := 0; i < l.count-l.index; i++ {
		fmt.Fprintf(l.w, "%s", moveDown)
	}
}

type MultiLogger struct {
	mu      sync.Mutex
	loggers []LineLogger
}

func New(count int, w io.Writer) MultiLogger {
	for i := 0; i < count; i++ {
		fmt.Fprintf(w, "hey\n")
	}
	loggers := make([]LineLogger, count)
	for i := 0; i < count; i++ {
		loggers[i] = LineLogger{index: i, count: count, w: w}
	}
	return MultiLogger{
		loggers: loggers,
	}
}

func (ml *MultiLogger) Printf(index int, format string, a ...any) {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	ml.loggers[index].Printf(format, a...)
}
