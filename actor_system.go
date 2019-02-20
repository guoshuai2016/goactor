package goactor

import (
	"errors"
	"fmt"
	"sync"
	"time"

	queue "github.com/scryner/lfreequeue"
)

type ExitEvent int

type ActorSystem struct {
	router              Router
	deadLetterProcessor DeadLetterProcessor
	actors              map[string][]*innerActor
	lock                *sync.RWMutex
}

func (system *ActorSystem) AddActor(name string, actorImpl ActorInterface) (ok bool, err error) {
	system.lock.Lock()

	actor := &innerActor{
		actorImpl:  actorImpl,
		notifyChan: make(chan interface{}, 1),
		events:     queue.NewQueue(),
		name:       name,
		system:     system,
	}

	if actors, ok := system.actors[name]; ok {
		system.actors[name] = append(actors, actor)
	} else {
		actors = make([]*innerActor, 1)
		actors[0] = actor
		system.actors[name] = actors
	}

	system.lock.Unlock()
	go actor.loop()
	return true, nil
}

func (system *ActorSystem) RemoveActor(name string, actorImpl ActorInterface) (ok bool, err error) {
	system.lock.Lock()
	defer system.lock.Unlock()
	if actors, ok := system.actors[name]; ok {
		if len(actors) == 1 && actors[0].actorImpl == actorImpl {
			actors[0].push(&Event{
				event:        ExitEvent(0),
				responseChan: nil,
			})
			delete(system.actors, name)
		} else {
			index := -1
			for i, actress := range actors {
				if actress.actorImpl == actorImpl {
					index = i
					break
				}
			}
			if index >= 0 {
				actors[index].push(&Event{
					event:        ExitEvent(0),
					responseChan: nil,
				})
				system.actors[name] = append(actors[:index], actors[index+1:]...)
				return true, nil
			}
		}
	}
	return false, errors.New("actor not in system")
}

func (system *ActorSystem) Shutdown() (ok bool, err error) {
	system.lock.Lock()
	defer system.lock.Unlock()
	for _, actors := range system.actors {
		for _, actor := range actors {
			actor.push(&Event{
				event:        ExitEvent(0),
				responseChan: nil,
			})
		}
	}
	// force clean
	system.actors = make(map[string][]*innerActor)
	return
}

func (system *ActorSystem) route(actorName string) (actor *innerActor, err error) {
	system.lock.RLock()
	defer system.lock.RUnlock()
	return system.router.Route(actorName, system.actors)
}

func (system *ActorSystem) require(actorName string, event interface{}, timeout int) (rst interface{}, err error) {
	if actor, err := system.route(actorName); err != nil {
		system.deadLetterProcessor.Process(actorName, event)
		return nil, err
	} else {
		var ch chan interface{}
		if timeout == -1 {
			ch = nil
		} else {
			ch = make(chan interface{}, 1)
		}

		actor.push(&Event{
			event:        event,
			responseChan: ch,
		})

		if timeout >= 0 {
			select {
			case rst = <-ch:
				return rst, nil
			case <-time.After(time.Duration(timeout) * time.Millisecond):
				return nil, errors.New(fmt.Sprintf("require to %s timeout", actorName))
			}
		} else {
			return nil, nil
		}
	}
}

func (system *ActorSystem) Request(actorName string, event interface{}) {
	_, _ = system.require(actorName, event, -1)
}

func (system *ActorSystem) Require(actorName string, event interface{}, timeoutInMilliSec int) (rst interface{}, err error) {
	return system.require(actorName, event, timeoutInMilliSec)
}

func NewDefaultActorSystem() *ActorSystem {
	return &ActorSystem{
		router:              NewFullQualifiedNameWithRandomBalancerRouter(),
		deadLetterProcessor: NewConsoleDeadLetterProcessor(),
		actors:              make(map[string][]*innerActor),
		lock:                &sync.RWMutex{},
	}
}
