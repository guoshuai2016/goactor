package system

import (
	"math/rand"
	"time"
)

type Balancer interface {
	Choose(actorName string, actors []*innerActor) *innerActor
}

type RandomBalancer struct {
	rand *rand.Rand
}

func NewRandomBalancer() Balancer {
	s1 := rand.NewSource(time.Now().UnixNano())
	return &RandomBalancer{
		rand: rand.New(s1),
	}
}

func (balancer RandomBalancer) Choose(actorName string, actors []*innerActor) *innerActor {
	index := balancer.rand.Intn(len(actors))
	return actors[index]
}
