package watcher

import "github.com/MV7VM/url-shortener/internal/domain/url-shortener/entities"

type observer interface {
	Update(*entities.Event)
}

// Watcher implements a simple publish/subscribe hub for audit events.
// Observers are notified sequentially in the order of registration.
type Watcher struct {
	observers []observer
}

// NewWatcher constructs an event watcher with an optional initial set of observers.
func NewWatcher(observers ...observer) *Watcher {
	return &Watcher{
		observers: observers,
	}
}

// Register adds a new observer that will receive future events.
func (e *Watcher) Register(o observer) {
	if e.observers == nil {
		e.observers = make([]observer, 0, 8)
	}
	e.observers = append(e.observers, o)
}

// Notify delivers the provided event to all registered observers.
func (e *Watcher) Notify(b *entities.Event) {
	for _, observer := range e.observers {
		observer.Update(b)
	}
}
