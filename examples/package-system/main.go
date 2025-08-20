package main

import (
	"fmt"

	"github.com/samber/do/v2"
	"github.com/samber/do/v2/examples/package-system/application"
	"github.com/samber/do/v2/examples/package-system/database"
	"github.com/samber/do/v2/examples/package-system/logging"
	"github.com/samber/do/v2/examples/package-system/services"
)

// Configuration represents application configuration
type Configuration struct {
	AppName string
	Port    int
	Debug   bool
}

func (c *Configuration) GetAppName() string {
	return c.AppName
}

func (c *Configuration) GetDebug() bool {
	return c.Debug
}

func main() {
	// Create injector with all packages
	injector := do.New(
		// Configuration is provided directly
		func(i do.Injector) {
			do.ProvideValue(i, &Configuration{
				AppName: "MyApp",
				Port:    8080,
				Debug:   true,
			})
		},
		// Apply modular packages
		database.DatabasePackage,
		logging.LoggingPackage,
		services.ServicesPackage,
		application.ApplicationPackage,
	)

	// Create aliases for Configuration to bridge interface types
	// This is needed for do.InvokeAs to work across packages
	do.As[*Configuration, database.Configuration](injector)
	do.As[*Configuration, logging.Configuration](injector)
	do.As[*Configuration, application.Configuration](injector)

	fmt.Println("=== Package System Example ===")
	fmt.Println("Available services:", injector.ListProvidedServices())

	// Run the application
	app := do.MustInvoke[*application.Application](injector)
	app.Run()

	fmt.Println("\n=== Package Benefits ===")
	fmt.Println("1. Modular service registration")
	fmt.Println("2. Reusable service packages")
	fmt.Println("3. Clear separation of concerns")
	fmt.Println("4. Easy testing and mocking")
}
