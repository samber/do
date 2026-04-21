package main

import (
	"fmt"

	"github.com/samber/do/v2"
)

func main() {
	injector := do.New()

	fmt.Println("=== Event-Driven Architecture Example ===")
	fmt.Println("This example demonstrates how to build an event-driven system")
	fmt.Printf("using the do library for dependency injection.\n\n")

	// Step 1: Register core infrastructure services
	fmt.Println("Step 1: Registering core infrastructure services")
	do.Provide(injector, func(i do.Injector) (*Logger, error) {
		return &Logger{Level: "INFO"}, nil
	})

	do.Provide(injector, func(i do.Injector) (*EventBus, error) {
		return NewEventBus(), nil
	})

	// Step 2: Register event handlers
	fmt.Println("Step 2: Registering event handlers")
	do.Provide(injector, func(i do.Injector) (*UserEventHandler, error) {
		return &UserEventHandler{
			Logger: do.MustInvoke[*Logger](i),
		}, nil
	})

	do.Provide(injector, func(i do.Injector) (*OrderEventHandler, error) {
		return &OrderEventHandler{
			Logger: do.MustInvoke[*Logger](i),
		}, nil
	})

	do.Provide(injector, func(i do.Injector) (*PaymentEventHandler, error) {
		return &PaymentEventHandler{
			Logger: do.MustInvoke[*Logger](i),
		}, nil
	})

	// Step 3: Register business services that publish events
	fmt.Println("Step 3: Registering business services that publish events")
	do.Provide(injector, func(i do.Injector) (*UserService, error) {
		return &UserService{
			EventBus: do.MustInvoke[*EventBus](i),
			Logger:   do.MustInvoke[*Logger](i),
		}, nil
	})

	do.Provide(injector, func(i do.Injector) (*OrderService, error) {
		return &OrderService{
			EventBus: do.MustInvoke[*EventBus](i),
			Logger:   do.MustInvoke[*Logger](i),
		}, nil
	})

	do.Provide(injector, func(i do.Injector) (*PaymentService, error) {
		return &PaymentService{
			EventBus: do.MustInvoke[*EventBus](i),
			Logger:   do.MustInvoke[*Logger](i),
		}, nil
	})

	// Step 4: Register main application
	fmt.Println("Step 4: Registering main application")
	do.Provide(injector, func(i do.Injector) (*Application, error) {
		return &Application{
			EventBus:       do.MustInvoke[*EventBus](i),
			UserService:    do.MustInvoke[*UserService](i),
			OrderService:   do.MustInvoke[*OrderService](i),
			PaymentService: do.MustInvoke[*PaymentService](i),
			Logger:         do.MustInvoke[*Logger](i),
		}, nil
	})

	fmt.Println("\n=== Setting up Event Handlers ===")
	fmt.Println("Subscribing event handlers to the event bus:")

	// Get the event bus and register handlers
	eventBus := do.MustInvoke[*EventBus](injector)
	userHandler := do.MustInvoke[*UserEventHandler](injector)
	orderHandler := do.MustInvoke[*OrderEventHandler](injector)
	paymentHandler := do.MustInvoke[*PaymentEventHandler](injector)

	// Subscribe handlers to events
	fmt.Println("  - UserEventHandler subscribed to 'user.created' events")
	eventBus.Subscribe(userHandler)

	fmt.Println("  - OrderEventHandler subscribed to 'order.created' events")
	eventBus.Subscribe(orderHandler)

	fmt.Println("  - PaymentEventHandler subscribed to 'payment.processed' events")
	eventBus.Subscribe(paymentHandler)

	fmt.Println("Event handlers registered successfully")

	fmt.Println("\n=== Service Information ===")
	fmt.Println("Available services:", injector.ListProvidedServices())

	// Run the application
	fmt.Println("\n=== Running Application ===")
	fmt.Println("Simulating a complete user journey with event publishing:")
	app := do.MustInvoke[*Application](injector)
	app.Run()

	fmt.Println("\n=== Event-Driven Architecture Benefits ===")
	fmt.Println("1. Loose coupling - services don't directly depend on each other")
	fmt.Println("2. Scalable event processing - easy to add new handlers")
	fmt.Println("3. Asynchronous processing - events can be processed independently")
	fmt.Println("4. Clear separation of concerns - publishers vs subscribers")
	fmt.Println("5. Testability - easy to mock event bus for testing")
	fmt.Println("6. Extensibility - new features can be added via event handlers")
}
