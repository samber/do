package main

import (
	"fmt"
)

// Application represents the main application.
type Application struct {
	EventBus       *EventBus
	UserService    *UserService
	OrderService   *OrderService
	PaymentService *PaymentService
	Logger         *Logger
}

func (app *Application) Run() {
	app.Logger.Log("Starting event-driven application...")

	// Simulate a user journey
	app.Logger.Log("=== User Journey Simulation ===")

	// Step 1: Create a user
	app.Logger.Log("Step 1: Creating user")
	if err := app.UserService.CreateUser("john_doe", "john@example.com"); err != nil {
		app.Logger.Log(fmt.Sprintf("Error creating user: %v", err))
		return
	}

	// Step 2: Create an order
	app.Logger.Log("Step 2: Creating order")
	if err := app.OrderService.CreateOrder("user-123", 99.99); err != nil {
		app.Logger.Log(fmt.Sprintf("Error creating order: %v", err))
		return
	}

	// Step 3: Process payment
	app.Logger.Log("Step 3: Processing payment")
	if err := app.PaymentService.ProcessPayment("order-123", 99.99); err != nil {
		app.Logger.Log(fmt.Sprintf("Error processing payment: %v", err))
		return
	}

	app.Logger.Log("User journey completed successfully!")

	// Show benefits
	app.Logger.Log("=== Event-Driven Architecture Benefits ===")
	app.Logger.Log("1. Loose coupling between services")
	app.Logger.Log("2. Scalable event processing")
	app.Logger.Log("3. Easy to add new event handlers")
	app.Logger.Log("4. Asynchronous event processing")
	app.Logger.Log("5. Clear separation of concerns")
}
