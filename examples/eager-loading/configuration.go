package main

import (
	"time"
)

// Configuration represents application configuration
type Configuration struct {
	AppName   string
	Port      int
	Debug     bool
	CreatedAt time.Time
}
