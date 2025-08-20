package logging

import (
	"fmt"
)

// Logger represents a logging service
type Logger struct {
	Level string
}

func (l *Logger) Log(message string) {
	fmt.Printf("[%s] %s\n", l.Level, message)
}
