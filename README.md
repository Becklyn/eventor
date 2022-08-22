# eventor

🔮 A minimalistic library for abstracting pub/sub operations

&rarr; *eventor* is [clerk](https://github.com/Becklyn/clerk) for pub/sub 😉

## Installation 

```sh
go get github.com/Becklyn/eventor
```

## Supported brokers

*eventor* has builtin support for the following brokers: 

- [Dapr Pub/sub API](https://docs.dapr.io/reference/api/pubsub_api/) - APIs for building portable and reliable microservices

## Usage

Being a minimalistic libary, *eventor* only provides you with the basiscs. The rest is up to your specific need.

### Creating a publisher 

```go
publisher := dapr.NewDaprPublisher(dapr.DaprPublisherConfig{
    Hostname: "localhost",
    Port:     "3500",
    Pubsub:   "redis-pubsub",
})
```

### Creating a subscriber registry

```go
subscriber := dapr.NewDaprSubscriberRegistry(dapr.DaprSubscriberRegistryConfig{
    App:    app,
    Pubsub: "redis-pubsub",
})
```

### Publish

```go
type Message struct {
    Id string `json:"id"`
    Body string `json:"body"`
}

if err := publisher.Publish("topic", Message{
    Id:   "0",
    Body: "Hello World",
}); err != nil {
    panic(err)
}
```

### Subscribe with handler function

```go
type Message struct {
    Id string `json:"id"`
    Body string `json:"body"`
}

handlerFn := func(msg Message) error {
    fmt.Println(msg)
}

usubscribe, err := eventor.On(handlerFn, subscriber, "topic")
if err != nil {
    panic(err)
}
defer unsubscribe()
```

## Subscribe with channel

```go
type Message struct {
    Id string `json:"id"`
    Body string `json:"body"`
}

messageChan, unsubscribe, err := eventor.Chan[Message](subscriber, "topic")
if err != nil {
    panic(err)
}
defer unsubscribe()

message := <-messageChan
fmt.Println(message)
```