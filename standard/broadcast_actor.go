package standard

import (
	. "github.com/goactor/system"
)

type BroadcastActor struct {
	BroadCastGroup []string // actor names
}

func (broadcaster *BroadcastActor) Receive(system *ActorSystem, eventType EventType, event interface{}) interface{} {
	if eventType != EVENT_REQUEST {
		println("BroadcastActor doesn't support event type \"%d\". Event will still be broadcast however.", eventType)
	}

	for _, name := range broadcaster.BroadCastGroup {
		system.Request(name, event)
	}

	return nil
}
