package goactor

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
	OnPlugin(system *ActorSystem)
	Receive(system *ActorSystem, eventType EventType, event interface{}) interface{}
	OnPullout(system *ActorSystem)
}

type innerActor struct {
	actorImpl  ActorInterface
	notifyChan chan *Event
	events     *queue.Queue

	name   string
	system *ActorSystem
}

func (actor *innerActor) push(event *Event) {
	actor.events.Enqueue(event)
	select {
	case actor.notifyChan <- nil:
	default:
	}
}

func (actor *innerActor) plugin() {
	actor.actorImpl.OnPlugin(actor.system)
}

func (actor *innerActor) pullout() {
	actor.actorImpl.OnPullout(actor.system)
}

func (actor *innerActor) loop() {
	actor.actorImpl.OnPlugin(actor.system)
	defer actor.actorImpl.OnPullout(actor.system)
	for {
		<-actor.notifyChan

		for {
			if event, ok := actor.events.Dequeue(); ok {
				typedEvent := event.(*Event)

				if _, ok := typedEvent.event.(ExitEvent); ok {
					// TODO: custom exiting!
					return
				}

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
