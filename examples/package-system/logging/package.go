package logging

import (
	"github.com/samber/do/v2"
)

// LoggingPackage is the global package for logging-related services.
var LoggingPackage = do.Package(
	do.Lazy(func(i do.Injector) (*Logger, error) {
		config := do.MustInvokeAs[Configuration](i)
		level := "INFO"
		if config.GetDebug() {
			level = "DEBUG"
		}
		return &Logger{Level: level}, nil
	}),
)
