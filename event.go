// Package wiston provides a robust event bus implementation.
package wiston

import (
	"errors"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// EventListener defines a map of subscription IDs to their corresponding callback functions.
type EventListener map[int64]func(data any)

// job represents a unit of work to be processed by a worker.
// It contains the event name and its associated data.
type job struct {
	name string
	data any
}

// PublishMode defines the behavior for the Publish method when the event queue is full.
type PublishMode int

const (
	// DropIfFull drops the event immediately if the queue is full.
	DropIfFull PublishMode = iota
	// BlockIfFull blocks the call until there is space in the queue.
	BlockIfFull
	// TimeoutIfFull waits for a specified duration for space to become available
	// before dropping the event.
	TimeoutIfFull
)

// EventBus provides a thread-safe, asynchronous event bus system.
// It allows for subscribing to events, publishing events, and unsubscribing.
// Events are processed concurrently by a pool of workers.
type EventBus struct {
	mu      *sync.RWMutex
	events  map[string]EventListener
	queue   chan job
	workers int
	wg      *sync.WaitGroup
	quit    chan struct{}
	counter int64
}

// NewEventBus creates and initializes a new EventBus with a specified number
// of workers and queue size. It also starts the worker pool.
func NewEventBus(workers int, queueSize int) *EventBus {
	bus := &EventBus{
		mu:      &sync.RWMutex{},
		events:  make(map[string]EventListener),
		queue:   make(chan job, queueSize),
		workers: workers,
		wg:      &sync.WaitGroup{},
		quit:    make(chan struct{}),
	}
	bus.startWorkers()
	return bus
}

// startWorkers launches the worker goroutines that process jobs from the queue.
func (b *EventBus) startWorkers() {
	for i := 0; i < b.workers; i++ {
		b.wg.Add(1)
		go func() {
			defer b.wg.Done()
			for {
				select {
				case j := <-b.queue:
					b.dispatch(j)
				case <-b.quit:
					return
				}
			}
		}()
	}
}

// dispatch finds all listeners for a given job's event and invokes them concurrently.
// It recovers from panics within listeners to prevent crashing the worker.
func (b *EventBus) dispatch(j job) {
	b.mu.RLock()
	listeners := make([]func(any), 0, len(b.events[j.name]))
	for _, fn := range b.events[j.name] {
		listeners = append(listeners, fn)
	}
	b.mu.RUnlock()

	for _, fn := range listeners {
		go func(fn func(any), data any) {
			defer func() {
				if r := recover(); r != nil {
					log.Println("listener panic:", r)
				}
			}()
			fn(data)
		}(fn, j.data)
	}
}

// Subscribe registers a callback function for a given event name.
// It returns a unique subscription ID that can be used to unsubscribe later.
func (b *EventBus) Subscribe(name string, callback func(any)) int64 {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.events[name] == nil {
		b.events[name] = make(EventListener)
	}

	id := atomic.AddInt64(&b.counter, 1)
	b.events[name][id] = callback
	return id
}

// Unsubscribe removes a subscription for a given event name and subscription ID.
func (b *EventBus) Unsubscribe(name string, id int64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if listeners, ok := b.events[name]; ok {
		delete(listeners, id)
	}
}

// Publish sends an event with the given name and data to the event bus.
// The behavior when the queue is full is determined by the PublishMode.
// An error is returned if the event is dropped or times out.
func (b *EventBus) Publish(name string, data any, mode PublishMode, timeout ...time.Duration) error {
	job := job{name, data}
	switch mode {
	case DropIfFull:
		select {
		case b.queue <- job:
		default:
			return errors.New("queue full: dropped event")
		}
	case BlockIfFull:
		b.queue <- job
	case TimeoutIfFull:
		if len(timeout) == 0 {
			return errors.New("timeout mode requires duration")
		}
		select {
		case b.queue <- job:
		case <-time.After(timeout[0]):
			return errors.New("queue full: publish timeout")
		}
	}
	return nil
}

// Close gracefully shuts down the EventBus.
// It stops all worker goroutines and waits for them to finish their current tasks.
func (b *EventBus) Close() {
	close(b.quit)
	b.wg.Wait()
}
