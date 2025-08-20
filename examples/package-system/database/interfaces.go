package database

// Configuration interface for dependency injection
type Configuration interface {
	GetAppName() string
}
