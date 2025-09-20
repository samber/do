package application

// Configuration interface for dependency injection.
type Configuration interface {
	GetAppName() string
	GetDebug() bool
}

// UserService interface for dependency injection.
type UserService interface {
	GetUser(id string) string
}

// OrderService interface for dependency injection.
type OrderService interface {
	GetOrder(id string) string
}

// Logger interface for dependency injection.
type Logger interface {
	Log(message string)
}
