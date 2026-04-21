package services

import (
	"fmt"
)

// UserService represents a user service.
type UserService struct {
	DB     Database
	Cache  Cache
	Logger Logger
}

func (u *UserService) GetUser(id string) string {
	u.Logger.Log(fmt.Sprintf("Getting user with ID: %s", id))

	// Try cache first
	if cached := u.Cache.Get("user:" + id); cached != nil {
		return fmt.Sprintf("Cached user: %v", cached)
	}

	// Query database
	result := u.DB.Query(fmt.Sprintf("SELECT * FROM users WHERE id = %s", id))

	// Cache the result
	u.Cache.Set("user:"+id, result)

	return result
}
