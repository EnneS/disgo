package main

type (
	Event interface {
		Name() string
	}

	EventHandler func(e Event)

	PlayEvent struct {
		Song Song
	}
	PauseEvent  struct{}
	ResumeEvent struct{}
	StopEvent   struct{}

	EventManager struct {
		eventQueue chan Event

		handlers map[string]EventHandler
	}
)

func (e *PlayEvent) Name() string {
	return "play"
}

func (e *PauseEvent) Name() string {
	return "pause"
}

func (e *ResumeEvent) Name() string {
	return "resume"
}

func (e *StopEvent) Name() string {
	return "stop"
}

// ==========================

func NewEventManager() *EventManager {
	em := &EventManager{
		eventQueue: make(chan Event),
		handlers:   make(map[string]EventHandler),
	}
	go em.Handle()
	return em
}

func (e *EventManager) Register(eventName string, handler EventHandler) {
	e.handlers[eventName] = handler
}

func (e *EventManager) Push(event Event) {
	e.eventQueue <- event
}

func (e *EventManager) Handle() {
	for {
		event := <-e.eventQueue
		switch event.(type) {
		case *PlayEvent:
			fn, ok := e.handlers["play"]
			if ok {
				fn(event)
			}
		case *PauseEvent:
			fn, ok := e.handlers["pause"]
			if ok {
				fn(event)
			}
		case *ResumeEvent:
			fn, ok := e.handlers["resume"]
			if ok {
				fn(event)
			}
		case *StopEvent:
			fn, ok := e.handlers["stop"]
			if ok {
				fn(event)
			}
		}
	}
}
