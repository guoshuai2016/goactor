package goactor

import (
	queue "github.com/scryner/lfreequeue"
	"testing"
	"time"
)

type testActor struct {
	plugin  bool
	message interface{}
	pullout bool
}

func (actor *testActor) OnPlugin(system *ActorSystem) {
	actor.plugin = true
}

func (actor *testActor) Receive(system *ActorSystem, eventType EventType, event interface{}) interface{} {
	actor.message = event
	return event
}

func (actor *testActor) OnPullout(system *ActorSystem) {
	actor.pullout = true
}

func TestActor(t *testing.T) {
	actorImpl := &testActor{false, nil, false}

	actor := &innerActor{
		actorImpl:  actorImpl,
		notifyChan: make(chan interface{}, 1),
		events:     queue.NewQueue(),
		name:       "na",
		system:     nil,
	}

	go actor.loop()

	// a moment for the actor initialize
	time.Sleep(time.Duration(10) * time.Millisecond)

	if actorImpl.plugin != true {
		t.Error("actor not plugged in")
	}

	actor.push(&Event{"test1", nil})

	// a moment for the actor process message
	time.Sleep(time.Duration(10) * time.Millisecond)

	if actorImpl.message != "test1" {
		t.Error("actor didn't get request message")
	}

	cn := make(chan interface{}, 1)
	actor.push(&Event{"test2", cn})
	if response := <-cn; response != "test2" {
		t.Error("actor didn't respond correct message")
	}
	if actorImpl.message != "test2" {
		t.Error("actor didn't get require message")
	}

	actor.push(&Event{ExitEvent(0), nil})

	// a moment for the actor deconstruct
	time.Sleep(time.Duration(10) * time.Millisecond)

	if actorImpl.pullout != true {
		t.Error("actor not pulled out")
	}
}
