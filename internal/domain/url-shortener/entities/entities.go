// Package entities содержит модели
// для сервиса сокращения ссылок.
package entities

import "sync"

// CtxKeyString is a helper type for context keys used inside the project.
type CtxKeyString string

const (
	// ActionShort is an audit action name used when a short URL is created.
	ActionShort = "shorten"
	// ActionFollow is an audit action name used when a short URL is followed.
	ActionFollow = "follow"
)

// Item represents a single shortened URL owned by a user.
type Item struct {
	UUID        string `json:"uuid,omitempty"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// BatchItem represents one element of a batch-shortening request/response.
type BatchItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url,omitempty"`
	ShortURL      string `json:"short_url"`
}

// Event describes a high-level user action that can be sent to auditors.
// generate:reset
type Event struct {
	TS     int    `json:"ts"`
	Action string `json:"action"`
	UserID string `json:"user_id"`
	URL    string `json:"url"`
}

// Resettable describes types that can reset their internal state.
type Resettable interface {
	Reset()
}

// Pool is a generic container for types that can be reset and reused.
type Pool[T Resettable] struct {
	pool sync.Pool
	new  func() T
}

// New creates and returns a new Pool for objects created by newFn.
func New[T Resettable](newFn func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{},
		new:  newFn,
	}
}

// Get returns an object from the pool or creates a new one if the pool is empty.
func (p *Pool[T]) Get() T {
	if v := p.pool.Get(); v != nil {
		return v.(T)
	}
	return p.new()
}

// Put resets the object and returns it to the pool.
func (p *Pool[T]) Put(v T) {
	if any(v) == nil {
		return
	}
	v.Reset()
	p.pool.Put(v)
}
