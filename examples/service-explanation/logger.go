package main

import (
	"fmt"
)

// Logger represents a logging service
type Logger struct {
	Config *Configuration
	Level  string
}

func (l *Logger) Log(message string) {
	fmt.Printf("[%s] %s: %s\n", l.Level, l.Config.AppName, message)
}
