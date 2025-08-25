package main

import (
	"fmt"
	"time"
)

// Logger represents a logging service.
type Logger struct {
	Level string
}

func (l *Logger) Log(message string) {
	fmt.Printf("[%s] %s\n", l.Level, message)
}

// UserService represents a user service.
type UserService struct {
	EventBus *EventBus
	Logger   *Logger
}

func (u *UserService) CreateUser(username, email string) error {
	userID := fmt.Sprintf("user-%d", time.Now().Unix())

	u.Logger.Log(fmt.Sprintf("Creating user: %s (%s)", username, email))

	// Publish user created event
	event := Event{
		Type: "user.created",
		Data: UserCreatedEvent{
			UserID:   userID,
			Username: username,
			Email:    email,
		},
		Timestamp: time.Now(),
	}

	return u.EventBus.Publish(event)
}

// OrderService represents an order service.
type OrderService struct {
	EventBus *EventBus
	Logger   *Logger
}

func (o *OrderService) CreateOrder(userID string, amount float64) error {
	orderID := fmt.Sprintf("order-%d", time.Now().Unix())

	o.Logger.Log(fmt.Sprintf("Creating order for user: %s, amount: %.2f", userID, amount))

	// Publish order created event
	event := Event{
		Type: "order.created",
		Data: OrderCreatedEvent{
			OrderID: orderID,
			UserID:  userID,
			Amount:  amount,
		},
		Timestamp: time.Now(),
	}

	return o.EventBus.Publish(event)
}

// PaymentService represents a payment service.
type PaymentService struct {
	EventBus *EventBus
	Logger   *Logger
}

func (p *PaymentService) ProcessPayment(orderID string, amount float64) error {
	paymentID := fmt.Sprintf("payment-%d", time.Now().Unix())

	p.Logger.Log(fmt.Sprintf("Processing payment for order: %s, amount: %.2f", orderID, amount))

	// Publish payment processed event
	event := Event{
		Type: "payment.processed",
		Data: PaymentProcessedEvent{
			PaymentID: paymentID,
			OrderID:   orderID,
			Amount:    amount,
			Status:    "completed",
		},
		Timestamp: time.Now(),
	}

	return p.EventBus.Publish(event)
}
