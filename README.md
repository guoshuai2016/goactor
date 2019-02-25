Actor Model for Go
==================

`goactor` is an prototype of [actor model](https://en.wikipedia.org/wiki/Actor_model) for Golang

While other actor models, such like Erlang, Akka and Vert.x, mainly target on performance, `goactor` target on decouple instead. Go language already ensure a good performance.

# Different decouple tech stacks
* Inverse of control (aka. Dependency injection)

Instead of rely on other components, a component rely on interface. It's until runtime when the real interface's implements are really "injected" into the components. Then it is possible for the IOC controller inject different implements in case of different runtime (eg. dev/stg/prd), thus the decouple.

In a word: A doesn't rely on B, A rely on IB. IB's implements varies among runtimes.

* EventBus

Each component does't rely on any other components, they rely only on EventBus. They publish events to EventBus, and retain interested events from EventBus. So there's totally no dependency except the EventBus.

* Actor model

Component not rely on others directly, they call others **by name**. Then it's possible to plug in different implements of **same name** in different runtimes.

---
As you can see, `goactor` is designed for event based system. Normally, for event based system, workflow is like:
1. Something __a__ happened to thread __A__
2. __A__ do something (eg. access DB, save a file, post a http request)
3. __A__ wait for another event to come

Then __A__ is heavily rely on SQL/File/Http module. It's not easy to mock a Http module, so __A__ is barely unable to UT.

In `goactor` way:
1. Something __a__ happened to thread __A__
2. __A__ require result from __http__ actor
3. __http__ actor give __A__ a result
4. __A__ post a message to __B__ actor (then __B__ do something)
5. __A__ wait for another event to come

Then __A__ just rely on names: __http__ and __B__

The implementation underneath of __http__ can be easily replaced.
That's the purpose of decouple.

---
# Performance

# Examples

## TODO