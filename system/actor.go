package system

import (
	queue "github.com/scryner/lfreequeue"
)

type Event struct {
	event        interface{}
	responseChan chan<- interface{}
}

type EventType int

const (
	EVENT_REQUIRE EventType = iota
	EVENT_REQUEST
)

type ActorInterface interface {
	Receive(system *ActorSystem, eventType EventType, event interface{}) interface{}
}

type innerActor struct {
	actorImpl  ActorInterface
	notifyChan chan *Event
	events     *queue.Queue

	name    string
	system  *ActorSystem
	exiting bool
}

func (actor *innerActor) push(event *Event) {
	actor.events.Enqueue(event)
	select {
	case actor.notifyChan <- nil:
	default:
	}
}

func (actor *innerActor) stop() {
	actor.exiting = true
	select {
	case actor.notifyChan <- nil:
	default:
	}
}

func (actor *innerActor) loop() {
	for {
		<-actor.notifyChan
		if actor.exiting {
			return
		}

		for {
			if event, ok := actor.events.Dequeue(); ok {
				typedEvent := event.(*Event)
				if typedEvent.responseChan == nil {
					actor.actorImpl.Receive(actor.system, EVENT_REQUEST, typedEvent.event)
				} else {
					typedEvent.responseChan <- actor.actorImpl.Receive(actor.system, EVENT_REQUIRE, typedEvent.event)
				}
			} else {
				break
			}
		}
	}
}
