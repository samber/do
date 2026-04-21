package main

import (
	"fmt"

	"github.com/samber/do/v2"
)

func main() {
	injector := do.New()

	fmt.Println("=== Service Explanation Example ===")
	fmt.Println("This example demonstrates how to use do.ExplainNamedService to")
	fmt.Printf("understand service dependencies, lifecycle, and debugging information.\n\n")

	// Step 1: Register Configuration (base service)
	fmt.Println("Step 1: Registering Configuration (base service)")
	do.Provide(injector, func(i do.Injector) (*Configuration, error) {
		return &Configuration{
			AppName: "MyApp",
			Port:    8080,
			Debug:   true,
		}, nil
	})

	// Step 2: Register Database (depends on Configuration)
	fmt.Println("Step 2: Registering Database (depends on Configuration)")
	do.Provide(injector, func(i do.Injector) (*Database, error) {
		config := do.MustInvoke[*Configuration](i)
		return &Database{
			Config: config,
			URL:    fmt.Sprintf("postgres://localhost:5432/%s", config.AppName),
		}, nil
	})

	// Step 3: Register Logger (depends on Configuration)
	fmt.Println("Step 3: Registering Logger (depends on Configuration)")
	do.Provide(injector, func(i do.Injector) (*Logger, error) {
		config := do.MustInvoke[*Configuration](i)
		level := "INFO"
		if config.Debug {
			level = "DEBUG"
		}
		return &Logger{
			Config: config,
			Level:  level,
		}, nil
	})

	// Step 4: Register Cache (independent service)
	fmt.Println("Step 4: Registering Cache (independent service)")
	do.Provide(injector, func(i do.Injector) (*Cache, error) {
		return &Cache{
			Data: make(map[string]interface{}),
		}, nil
	})

	// Step 5: Register Business Services (depend on multiple services)
	fmt.Println("Step 5: Registering Business Services (depend on multiple services)")
	do.Provide(injector, func(i do.Injector) (*UserService, error) {
		return &UserService{
			DB:     do.MustInvoke[*Database](i),
			Cache:  do.MustInvoke[*Cache](i),
			Logger: do.MustInvoke[*Logger](i),
		}, nil
	})

	do.Provide(injector, func(i do.Injector) (*OrderService, error) {
		return &OrderService{
			DB:     do.MustInvoke[*Database](i),
			Cache:  do.MustInvoke[*Cache](i),
			Logger: do.MustInvoke[*Logger](i),
		}, nil
	})

	// Step 6: Register Application (depends on all services)
	fmt.Println("Step 6: Registering Application (depends on all services)")
	do.Provide(injector, func(i do.Injector) (*Application, error) {
		return &Application{
			Config:       do.MustInvoke[*Configuration](i),
			DB:           do.MustInvoke[*Database](i),
			Cache:        do.MustInvoke[*Cache](i),
			Logger:       do.MustInvoke[*Logger](i),
			UserService:  do.MustInvoke[*UserService](i),
			OrderService: do.MustInvoke[*OrderService](i),
		}, nil
	})

	fmt.Println("\n=== Service Information ===")
	fmt.Println("Available services:", injector.ListProvidedServices())

	fmt.Println("\n=== Service Explanations (Before Invocation) ===")
	fmt.Println("Explaining each service before any invocations:")

	// Explain each service before invocation
	services := []string{
		"*main.Configuration",
		"*main.Database",
		"*main.Logger",
		"*main.Cache",
		"*main.UserService",
		"*main.OrderService",
		"*main.Application",
	}

	for _, serviceName := range services {
		fmt.Printf("\n--- Explaining %s ---\n", serviceName)
		explanation, found := do.ExplainNamedService(injector, serviceName)
		if found {
			fmt.Println(explanation.String())
		} else {
			fmt.Println("Service not found")
		}
	}

	fmt.Println("\n=== Dependency Graph Analysis ===")
	fmt.Println("Analyzing dependency relationships:")

	// Show dependency relationships
	fmt.Println("\nConfiguration (root service):")
	configExplanation, found := do.ExplainNamedService(injector, "*main.Configuration")
	if found {
		fmt.Println(configExplanation.String())
	}

	fmt.Println("\nDatabase (depends on Configuration):")
	dbExplanation, found := do.ExplainNamedService(injector, "*main.Database")
	if found {
		fmt.Println(dbExplanation.String())
	}

	fmt.Println("\nUserService (depends on Database, Cache, Logger):")
	userServiceExplanation, found := do.ExplainNamedService(injector, "*main.UserService")
	if found {
		fmt.Println(userServiceExplanation.String())
	}

	fmt.Println("\n=== Running Application ===")
	fmt.Println("Invoking the Application service to trigger all dependencies:")
	app := do.MustInvoke[*Application](injector)
	app.Run()

	fmt.Println("\n=== Post-Execution Explanations ===")
	fmt.Println("Explaining services after they have been invoked:")

	// Show explanations after services have been invoked
	fmt.Println("\nApplication (after invocation):")
	appExplanation, found := do.ExplainNamedService(injector, "*main.Application")
	if found {
		fmt.Println(appExplanation.String())
	}

	fmt.Println("\n=== Service Explanation Benefits ===")
	fmt.Println("1. Service dependency visualization - understand what depends on what")
	fmt.Println("2. Service lifecycle tracking - see when services are created and invoked")
	fmt.Println("3. Invocation source tracking - know which service triggered the creation")
	fmt.Println("4. Service type information - understand the concrete types being used")
	fmt.Println("5. Build time measurements - see how long service creation takes")
	fmt.Println("6. Debugging support - identify dependency issues and circular references")
}
