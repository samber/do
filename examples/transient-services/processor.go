package main

import (
	"fmt"
)

// OrderProcessor represents the main service that coordinates order processing.
type OrderProcessor struct {
	UserService    *UserService
	OrderService   *OrderService
	PaymentService *PaymentService
}

func (op *OrderProcessor) ProcessOrder() {
	fmt.Println("=== Processing Order ===")
	op.UserService.GetUserInfo()
	op.OrderService.CreateOrder()
	op.PaymentService.ProcessPayment()
	fmt.Println("=== Order Processed ===")
}

// RequestHandler represents a handler for HTTP requests.
type RequestHandler struct {
	Processor *OrderProcessor
}

func (h *RequestHandler) HandleRequest(userID, session string) {
	fmt.Printf("\n--- Handling request for user: %s, session: %s ---\n", userID, session)
	h.Processor.ProcessOrder()
}
