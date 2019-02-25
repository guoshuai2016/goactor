package goactor

import (
	"testing"
)

type singleBalancer struct{}

func (balancer *singleBalancer) Choose(actorName string, actors []*innerActor) *innerActor {
	return actors[0]
}

func TestFullQualifiedNameRouter(t *testing.T) {
	router := NewFullQualifiedNameWithCustomBalancerRouter(&singleBalancer{})

	actors := make(map[string][]*innerActor)

	actors["A"] = []*innerActor{&innerActor{name: "A"}}
	actors["B"] = []*innerActor{&innerActor{name: "B"}}
	actors["C"] = []*innerActor{&innerActor{name: "C"}}
	actors["D"] = []*innerActor{&innerActor{name: "D"}}

	if actor, err := router.Route("A", actors); actor.name != "A" || err != nil {
		t.Error("Cannot rout to A")
	}

	if actor, err := router.Route("B", actors); actor.name != "B" || err != nil {
		t.Error("Cannot rout to B")
	}

	if actor, err := router.Route("C", actors); actor.name != "C" || err != nil {
		t.Error("Cannot rout to C")
	}

	if actor, err := router.Route("D", actors); actor.name != "D" || err != nil {
		t.Error("Cannot rout to D")
	}

	if actor, err := router.Route("E", actors); actor != nil || err == nil {
		t.Error("Unexpected E found")
	}
}
