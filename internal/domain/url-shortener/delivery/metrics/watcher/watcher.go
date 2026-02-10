package watcher

import "github.com/MV7VM/url-shortener/internal/domain/url-shortener/entities"

type observer interface {
	Update(*entities.Event)
}

// реализация publisher
type Watcher struct {
	observers []observer
}

func NewWatcher(observers ...observer) *Watcher {
	return &Watcher{
		observers: observers,
	}
}

func (e *Watcher) Register(o observer) {
	if e.observers == nil {
		e.observers = make([]observer, 0, 8)
	}
	e.observers = append(e.observers, o)
}

func (e *Watcher) Notify(b *entities.Event) {
	for _, observer := range e.observers {
		observer.Update(b)
	}
}
