package logging

// Configuration interface for dependency injection
type Configuration interface {
	GetDebug() bool
}
