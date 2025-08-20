package main

import (
	"time"
)

// Event represents a generic event
type Event struct {
	Type      string
	Data      interface{}
	Timestamp time.Time
}

// UserCreatedEvent represents a user created event
type UserCreatedEvent struct {
	UserID   string
	Username string
	Email    string
}

// OrderCreatedEvent represents an order created event
type OrderCreatedEvent struct {
	OrderID string
	UserID  string
	Amount  float64
}

// PaymentProcessedEvent represents a payment processed event
type PaymentProcessedEvent struct {
	PaymentID string
	OrderID   string
	Amount    float64
	Status    string
}
