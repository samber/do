package application

import (
	"fmt"
)

// Application represents the main application.
type Application struct {
	Config       Configuration
	UserService  UserService
	OrderService OrderService
	Logger       Logger
}

func (app *Application) Run() {
	app.Logger.Log("Starting application...")

	// Test user service
	fmt.Println("User data:", app.UserService.GetUser("123"))
	fmt.Println("User data (cached):", app.UserService.GetUser("123"))

	// Test order service
	fmt.Println("Order data:", app.OrderService.GetOrder("456"))
	fmt.Println("Order data (cached):", app.OrderService.GetOrder("456"))

	app.Logger.Log("Application running successfully")
}
