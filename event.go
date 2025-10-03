package wiston

import (
	"errors"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type EventListener map[int64]func(data any)

type job struct {
	name string
	data any
}

type PublishMode int

const (
	DropIfFull PublishMode = iota
	BlockIfFull
	TimeoutIfFull
)

type EventBus struct {
	mu      *sync.RWMutex
	events  map[string]EventListener
	queue   chan job
	workers int
	wg      *sync.WaitGroup
	quit    chan struct{}
	counter int64
}

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

func (b *EventBus) Unsubscribe(name string, id int64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if listeners, ok := b.events[name]; ok {
		delete(listeners, id)
	}
}

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

func (b *EventBus) Close() {
	close(b.quit)
	b.wg.Wait()
}
