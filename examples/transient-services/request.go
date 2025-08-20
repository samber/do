package main

import (
	"time"
)

// RequestID represents a unique identifier for each request
type RequestID struct {
	ID        string
	CreatedAt time.Time
}

// RequestContext represents context for each request
type RequestContext struct {
	ID      *RequestID
	UserID  string
	Session string
}
