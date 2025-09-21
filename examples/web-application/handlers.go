package main

import (
	"fmt"
	"net/http"
)

// UserHandler represents HTTP handler for user operations.
type UserHandler struct {
	UserService *UserService
	Logger      *Logger
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("id")
	if userID == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	h.Logger.Log(fmt.Sprintf("HTTP request: GET /users/%s", userID))

	userData := h.UserService.GetUser(userID)
	fmt.Fprintf(w, "User data: %s", userData) //nolint:errcheck
}

// OrderHandler represents HTTP handler for order operations.
type OrderHandler struct {
	OrderService *OrderService
	Logger       *Logger
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	orderID := r.URL.Query().Get("id")
	if orderID == "" {
		orderID = "456" // default
	}

	result := h.OrderService.GetOrder(orderID)
	fmt.Fprintf(w, "Order: %s", result) //nolint:errcheck
}

// HealthHandler represents HTTP handler for health checks.
type HealthHandler struct {
	DB     *Database
	Cache  *Cache
	Logger *Logger
}

func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Check database health
	if err := h.DB.HealthCheck(); err != nil {
		http.Error(w, "Database unhealthy", http.StatusServiceUnavailable)
		return
	}

	// Check cache health
	if err := h.Cache.HealthCheck(); err != nil {
		http.Error(w, "Cache unhealthy", http.StatusServiceUnavailable)
		return
	}

	// Check logger health
	if err := h.Logger.HealthCheck(); err != nil {
		http.Error(w, "Logger unhealthy", http.StatusServiceUnavailable)
		return
	}

	fmt.Fprintf(w, "All services healthy") //nolint:errcheck
}
