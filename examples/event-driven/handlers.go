package main

import (
	"fmt"
)

// UserEventHandler handles user-related events.
type UserEventHandler struct {
	Logger *Logger
}

func (h *UserEventHandler) Handle(event Event) error {
	h.Logger.Log(fmt.Sprintf("UserEventHandler: Processing event %s", event.Type))

	switch event.Type {
	case "user.created":
		if userData, ok := event.Data.(UserCreatedEvent); ok {
			h.Logger.Log(fmt.Sprintf("User created: ID=%s, Username=%s, Email=%s",
				userData.UserID, userData.Username, userData.Email))
		}
	case "user.updated":
		h.Logger.Log("User updated event received")
	}

	return nil
}

func (h *UserEventHandler) GetEventType() string {
	return "user.created"
}

// OrderEventHandler handles order-related events.
type OrderEventHandler struct {
	Logger *Logger
}

func (h *OrderEventHandler) Handle(event Event) error {
	h.Logger.Log(fmt.Sprintf("OrderEventHandler: Processing event %s", event.Type))

	switch event.Type {
	case "order.created":
		if orderData, ok := event.Data.(OrderCreatedEvent); ok {
			h.Logger.Log(fmt.Sprintf("Order created: ID=%s, UserID=%s, Amount=%.2f",
				orderData.OrderID, orderData.UserID, orderData.Amount))
		}
	case "order.cancelled":
		h.Logger.Log("Order cancelled event received")
	}

	return nil
}

func (h *OrderEventHandler) GetEventType() string {
	return "order.created"
}

// PaymentEventHandler handles payment-related events.
type PaymentEventHandler struct {
	Logger *Logger
}

func (h *PaymentEventHandler) Handle(event Event) error {
	h.Logger.Log(fmt.Sprintf("PaymentEventHandler: Processing event %s", event.Type))

	switch event.Type {
	case "payment.processed":
		if paymentData, ok := event.Data.(PaymentProcessedEvent); ok {
			h.Logger.Log(fmt.Sprintf("Payment processed: ID=%s, OrderID=%s, Amount=%.2f, Status=%s",
				paymentData.PaymentID, paymentData.OrderID, paymentData.Amount, paymentData.Status))
		}
	case "payment.failed":
		h.Logger.Log("Payment failed event received")
	}

	return nil
}

func (h *PaymentEventHandler) GetEventType() string {
	return "payment.processed"
}
