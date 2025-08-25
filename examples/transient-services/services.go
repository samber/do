package main

import (
	"fmt"
)

// UserService represents a service that handles user operations.
type UserService struct {
	Context *RequestContext
}

func (u *UserService) GetUserInfo() {
	fmt.Printf("Getting user info for user %s in session %s (request: %s)\n",
		u.Context.UserID, u.Context.Session, u.Context.ID.ID)
}

// OrderService represents a service that handles order operations.
type OrderService struct {
	Context *RequestContext
}

func (o *OrderService) CreateOrder() {
	fmt.Printf("Creating order for user %s in session %s (request: %s)\n",
		o.Context.UserID, o.Context.Session, o.Context.ID.ID)
}

// PaymentService represents a service that handles payment operations.
type PaymentService struct {
	Context *RequestContext
}

func (p *PaymentService) ProcessPayment() {
	fmt.Printf("Processing payment for user %s in session %s (request: %s)\n",
		p.Context.UserID, p.Context.Session, p.Context.ID.ID)
}
