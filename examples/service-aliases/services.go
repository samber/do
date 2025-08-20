package main

import (
	"fmt"
)

// AuditService represents a service that only needs read access
type AuditService struct {
	DB ReadOnlyDatabase
}

func (a *AuditService) AuditUserAccess(userID string) {
	result := a.DB.Query(fmt.Sprintf("SELECT access_log FROM users WHERE id = %s", userID))
	fmt.Printf("Audit result: %s\n", result)
}

// ConnectionManager represents a service that manages database connections
type ConnectionManager struct {
	DB WriteDatabase
}

func (cm *ConnectionManager) Initialize() error {
	return cm.DB.Connect()
}
