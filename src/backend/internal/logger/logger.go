package logger

import (
	"fmt"
	"sync"
	"time"
)

type Level string

const (
	INFO  Level = "INFO"
	WARN  Level = "WARN"
	ERROR Level = "ERR"
	SYS   Level = "SYS"
)

type Entry struct {
	Time    time.Time `json:"time"`
	Level   Level     `json:"level"`
	Message string    `json:"message"`
}

type Logger struct {
	mu        sync.RWMutex
	entries   []Entry
	maxSize   int
	listeners []chan Entry
}

func New(maxSize int) *Logger {
	return &Logger{
		maxSize: maxSize,
	}
}

func (l *Logger) log(level Level, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	entry := Entry{
		Time:    time.Now(),
		Level:   level,
		Message: msg,
	}

	l.mu.Lock()
	l.entries = append(l.entries, entry)
	if len(l.entries) > l.maxSize {
		l.entries = l.entries[len(l.entries)-l.maxSize:]
	}
	listeners := make([]chan Entry, len(l.listeners))
	copy(listeners, l.listeners)
	l.mu.Unlock()

	for _, ch := range listeners {
		select {
		case ch <- entry:
		default:
		}
	}

	fmt.Printf("[%s] [%s] %s\n", entry.Time.Format("15:04:05.000"), level, msg)
}

func (l *Logger) Info(format string, args ...any) {
	l.log(INFO, format, args...)
}

func (l *Logger) Warn(format string, args ...any) {
	l.log(WARN, format, args...)
}

func (l *Logger) Error(format string, args ...any) {
	l.log(ERROR, format, args...)
}

func (l *Logger) Sys(format string, args ...any) {
	l.log(SYS, format, args...)
}

func (l *Logger) Entries() []Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	cp := make([]Entry, len(l.entries))
	copy(cp, l.entries)
	return cp
}

func (l *Logger) Subscribe() chan Entry {
	ch := make(chan Entry, 64)
	l.mu.Lock()
	l.listeners = append(l.listeners, ch)
	l.mu.Unlock()
	return ch
}

func (l *Logger) Unsubscribe(ch chan Entry) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for i, listener := range l.listeners {
		if listener == ch {
			l.listeners = append(l.listeners[:i], l.listeners[i+1:]...)
			close(ch)
			return
		}
	}
}
