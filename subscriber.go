package eventor

import (
	"context"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

// HandlerFn is a function that is called when a cloudevent is received at a subscriber.
type HandlerFn func(event cloudevents.Event, parentSpanCtx context.Context) error

type Subscriber interface {
	// Subscribe registers a subscriber at the registry.
	// A topic needs to be passed from which the subscriber will receive events.
	// The handler function is called when an event is received.
	Subscribe(topic string, handler HandlerFn) error

	// Unsubscribe removes a subscriber from the registry.
	// The handler function of the handler to remove has to be passed.
	Unsubscribe(handler HandlerFn) error
}
