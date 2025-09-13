package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime/debug"
	"sync"
)

type Level int

const (
	LevelInfo Level = iota
	LevelWarn
	LevelError
)

type ErrorWithContext struct {
	Msg   string
	Err   error
	Stack []byte
}

func (e *ErrorWithContext) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s | %v", e.Msg, e.Err)
	}
	return e.Msg
}

type Logger struct {
	mu      sync.Mutex
	writers []io.Writer
	debug   bool
}

func New(writers ...io.Writer) *Logger {
	if len(writers) == 0 {
		writers = []io.Writer{os.Stdout}
	}
	return &Logger{
		writers: writers,
		debug:   false,
	}
}

func (l *Logger) EnableDebug(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.debug = enabled
}

func (l *Logger) Wrap(err error, msg string, args ...any) *ErrorWithContext {
	stack := []byte(nil)
	if l.debug {
		stack = debug.Stack()
	}

	return &ErrorWithContext{
		Msg:   msg,
		Err:   err,
		Stack: stack,
	}
}

func (l *Logger) log(level string, msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	logger := log.New(io.MultiWriter(l.writers...), level+": ", log.Ldate|log.Ltime)
	logger.Println(msg)
}

func (l *Logger) Info(msg string, args ...any) { l.log("INFO", fmt.Sprintf(msg, args...)) }
func (l *Logger) Warn(msg string, args ...any) { l.log("WARN", fmt.Sprintf(msg, args...)) }

func (l *Logger) Error(err error, msg string, args ...any) {
	fullMsg := fmt.Sprintf("%s | %v", fmt.Sprintf(msg, args...), err)
	ewc, ok := err.(*ErrorWithContext)
	if !ok {
		ewc = l.Wrap(err, msg, args...)
	}

	if len(ewc.Stack) > 0 {
		fullMsg += fmt.Sprintf("\nStack:\n%s", string(ewc.Stack))
	}

	l.log("ERROR", fullMsg)
}

func (l *Logger) Fatal(err error, msg string, args ...any) {
	l.Error(err, msg, args...)
	os.Exit(1)
}
