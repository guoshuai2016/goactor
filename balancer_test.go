package goactor

import "testing"

type mockActor string

func (actor *mockActor) OnPlugin(system *ActorSystem) {
}

func (actor *mockActor) Receive(system *ActorSystem, eventType EventType, event interface{}) interface{} {
	return event
}

func (actor *mockActor) OnPullout(system *ActorSystem) {
}

func TestBalancer(t *testing.T) {
	mocks := []mockActor{"A", "B", "C", "D"}

	actors := make([]*innerActor, len(mocks))
	count := make(map[string]int)

	for i, m := range mocks {
		actors[i] = &innerActor{
			actorImpl:  &m,
			notifyChan: nil,
			events:     nil,
			name:       string(m),
			system:     nil,
		}
		count[string(m)] = 0
	}

	balancer := NewRandomBalancer()

	for i := 0; i < 10000; i++ {
		actor := balancer.Choose("na", actors)
		count[actor.name] = count[actor.name] + 1
	}

	for _, m := range mocks {
		c := count[string(m)]
		if c >= 10000/len(mocks)+100 || c <= 10000/len(mocks)-100 {
			t.Error("RandomBalance doesn't have even distribution")
		}
	}
}
