package standard

import (
	. "github.com/xxpxxxxp/goactor"
)

type BroadcastActor struct {
	BroadCastGroup []string // actor names
}

func (broadcaster *BroadcastActor) OnPlugin(system *ActorSystem) {}

func (broadcaster *BroadcastActor) Receive(system *ActorSystem, eventType EventType, event interface{}) interface{} {
	if eventType != EVENT_REQUEST {
		println("BroadcastActor doesn't support event type \"%d\". Event will still be broadcast however.", eventType)
	}

	for _, name := range broadcaster.BroadCastGroup {
		system.Request(name, event)
	}

	return nil
}

func (broadcaster *BroadcastActor) OnPullout(system *ActorSystem) {}
