package goactor

import (
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"
)

type singleBenchmarkActor struct{}

func (actor *singleBenchmarkActor) OnPlugin(system *ActorSystem) {}
func (actor *singleBenchmarkActor) Receive(system *ActorSystem, eventType EventType, event interface{}) interface{} {
	// pretending like we are doing something
	time.Sleep(time.Millisecond)
	return event
}

func (actor *singleBenchmarkActor) OnPullout(system *ActorSystem) {}

func BenchmarkCallByActor(b *testing.B) {
	system := NewDefaultActorSystem()

	actor := &singleBenchmarkActor{}
	system.AddActor("benchmark", actor)
	for i := 0; i < b.N; i++ {
		system.Require("benchmark", 0, 1000)
	}

	system.Shutdown()
}

func BenchmarkCallDirect(b *testing.B) {
	system := NewDefaultActorSystem()

	actor := &singleBenchmarkActor{}
	system.AddActor("benchmark", actor)
	for i := 0; i < b.N; i++ {
		actor.Receive(system, EVENT_REQUEST, 0)
	}

	system.Shutdown()
}

type multiBenchmarkActor struct{}

func (actor *multiBenchmarkActor) OnPlugin(system *ActorSystem) {}
func (actor *multiBenchmarkActor) Receive(system *ActorSystem, eventType EventType, event interface{}) interface{} {
	time.Sleep(time.Microsecond)
	return event
}

func (actor *multiBenchmarkActor) OnPullout(system *ActorSystem) {}

func BenchmarkMultipleActors(b *testing.B) {
	system := NewDefaultActorSystem()
	system.AddActor("benchmark0", &multiBenchmarkActor{})
	system.AddActor("benchmark0", &multiBenchmarkActor{})
	system.AddActor("benchmark0", &multiBenchmarkActor{})
	system.AddActor("benchmark0", &multiBenchmarkActor{})
	system.AddActor("benchmark0", &multiBenchmarkActor{})
	system.AddActor("benchmark1", &multiBenchmarkActor{})
	system.AddActor("benchmark1", &multiBenchmarkActor{})
	system.AddActor("benchmark1", &multiBenchmarkActor{})
	system.AddActor("benchmark1", &multiBenchmarkActor{})
	system.AddActor("benchmark1", &multiBenchmarkActor{})
	system.AddActor("benchmark2", &multiBenchmarkActor{})
	system.AddActor("benchmark2", &multiBenchmarkActor{})
	system.AddActor("benchmark2", &multiBenchmarkActor{})
	system.AddActor("benchmark2", &multiBenchmarkActor{})
	system.AddActor("benchmark2", &multiBenchmarkActor{})
	system.AddActor("benchmark3", &multiBenchmarkActor{})
	system.AddActor("benchmark3", &multiBenchmarkActor{})
	system.AddActor("benchmark3", &multiBenchmarkActor{})
	system.AddActor("benchmark3", &multiBenchmarkActor{})
	system.AddActor("benchmark3", &multiBenchmarkActor{})
	system.AddActor("benchmark4", &multiBenchmarkActor{})
	system.AddActor("benchmark4", &multiBenchmarkActor{})
	system.AddActor("benchmark4", &multiBenchmarkActor{})
	system.AddActor("benchmark4", &multiBenchmarkActor{})
	system.AddActor("benchmark4", &multiBenchmarkActor{})

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	actorNames := []string{"benchmark0", "benchmark1", "benchmark2", "benchmark3", "benchmark4"}
	for i := 0; i < b.N; i++ {
		system.Require(actorNames[r.Intn(5)], 0, 1000)
	}

	system.Shutdown()
}

type functionTestActor struct {
	allActors []string
	r         *rand.Rand
	count     *int32
}

func (actor *functionTestActor) OnPlugin(system *ActorSystem) {}
func (actor *functionTestActor) Receive(system *ActorSystem, eventType EventType, event interface{}) interface{} {
	nextActor := actor.allActors[actor.r.Intn(len(actor.allActors))]

	var et EventType
	switch actor.r.Intn(4) {
	case 0:
		et = EVENT_REQUIRE
		system.Require(nextActor, event, 10)
	default:
		et = EVENT_REQUEST
		system.Request(nextActor, event)
	}

	atomic.AddInt32(actor.count, 1)
	fmt.Println(fmt.Sprintf("Calling next actor: %s, type: %d", nextActor, et))

	return event
}

func (actor *functionTestActor) OnPullout(system *ActorSystem) {}

func TestActorSystem(t *testing.T) {
	actorNames := []string{"benchmark0", "benchmark1", "benchmark2", "benchmark3", "benchmark4"}

	count := int32(0)

	system := NewDefaultActorSystem()
	for _, actorName := range actorNames {
		for i := 0; i < 5; i++ {
			system.AddActor(actorName,
				&functionTestActor{
					allActors: actorNames,
					r:         rand.New(rand.NewSource(time.Now().UnixNano())),
					count:     &count,
				})
		}
	}

	system.Request(actorNames[0], 0)

	time.Sleep(time.Duration(1) * time.Second)
	system.Shutdown()

	fmt.Println("Actor system total call count: ", count)
}
