package main

import (
	"fmt"
	"sync"
)

// EventHandler represents an event handler interface.
type EventHandler interface {
	Handle(event Event) error
	GetEventType() string
}

// EventBus represents an event bus for publishing and subscribing to events.
type EventBus struct {
	handlers map[string][]EventHandler
	mu       sync.RWMutex
}

func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[string][]EventHandler),
	}
}

func (eb *EventBus) Subscribe(handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eventType := handler.GetEventType()
	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
}

func (eb *EventBus) Publish(event Event) error {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	handlers, exists := eb.handlers[event.Type]
	if !exists {
		return fmt.Errorf("no handlers for event type: %s", event.Type)
	}

	for _, handler := range handlers {
		if err := handler.Handle(event); err != nil {
			return fmt.Errorf("handler error: %w", err)
		}
	}

	return nil
}
