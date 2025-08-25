package main

import (
	"fmt"
	"time"

	"github.com/samber/do/v2"
)

func main() {
	// Create injector with detailed logging for demonstration
	injector := do.NewWithOpts(&do.InjectorOpts{
		Logf: func(format string, args ...any) {
			fmt.Printf("[DI] "+format+"\n", args...)
		},
	})

	fmt.Println("=== Eager Loading Example ===")
	fmt.Println("Eager services are instantiated immediately when registered")
	fmt.Println("This is useful for services that need to be available immediately")
	fmt.Printf("or for services that perform initialization work on startup.\n\n")

	// Step 1: Register Configuration (eager)
	fmt.Println("Step 1: Registering Configuration (eager)")
	do.ProvideValue(injector, &Configuration{
		AppName:   "MyApp",
		Port:      8080,
		Debug:     true,
		CreatedAt: time.Now(),
	})

	// Step 2: Register Database (eager with dependency injection)
	fmt.Println("Step 2: Registering Database (eager with config dependency)")
	do.ProvideValue(injector, &Database{
		Config: do.MustInvoke[*Configuration](injector),
		URL:    "postgres://localhost:5432/mydb",
	})

	// Step 3: Register Logger (eager with dependency injection)
	fmt.Println("Step 3: Registering Logger (eager with config dependency)")
	do.ProvideValue(injector, &Logger{
		Config: do.MustInvoke[*Configuration](injector),
		Level:  "INFO",
	})

	fmt.Println("\n=== All Eager Services Registered ===")
	fmt.Println("Services in container:", injector.ListProvidedServices())

	// Step 4: Register Application (lazy - for comparison)
	fmt.Println("\nStep 4: Registering Application (lazy - for comparison)")
	do.Provide(injector, func(i do.Injector) (*Application, error) {
		return &Application{
			Config: do.MustInvoke[*Configuration](i),
			DB:     do.MustInvoke[*Database](i),
			Logger: do.MustInvoke[*Logger](i),
		}, nil
	})

	fmt.Println("\n=== Starting Application ===")
	app := do.MustInvoke[*Application](injector)
	app.Start()

	fmt.Println("\n=== Eager Loading Benefits ===")
	fmt.Println("1. Immediate availability - no waiting for first request")
	fmt.Println("2. Early error detection - issues found at startup")
	fmt.Println("3. Predictable resource usage - all services initialized upfront")
	fmt.Println("4. Better for critical services that must be ready immediately")

	fmt.Println("\n=== Service Information ===")
	fmt.Println("Configuration created at:", app.Config.CreatedAt)
	fmt.Println("Database URL:", app.DB.URL)
	fmt.Println("Logger level:", app.Logger.Level)
}
