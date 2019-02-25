package goactor

import (
	"errors"
	"fmt"
)

type Router interface {
	Route(actorName string, actors map[string][]*innerActor) (actor *innerActor, err error)
}

type FullQualifiedNameRouter struct {
	balancer Balancer
}

func NewFullQualifiedNameWithCustomBalancerRouter(balancer Balancer) *FullQualifiedNameRouter {
	return &FullQualifiedNameRouter{
		balancer: balancer,
	}
}

func NewFullQualifiedNameWithRandomBalancerRouter() *FullQualifiedNameRouter {
	return NewFullQualifiedNameWithCustomBalancerRouter(NewRandomBalancer())
}

func (router *FullQualifiedNameRouter) Route(actorName string, actors map[string][]*innerActor) (actor *innerActor, err error) {
	if _, ok := actors[actorName]; ok {
		return router.balancer.Choose(actorName, actors[actorName]), nil
	} else {
		return nil, errors.New(fmt.Sprintf("Unable to find actor match %s", actorName))
	}
}
