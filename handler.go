package eventor

import cloudevents "github.com/cloudevents/sdk-go/v2"

// UnsubscribeFn is a function to unsubscribe a handler.
type UnsubscribeFn func()

// On is a convenience function for creating a handler that executes only if the type can be parsed.
// A subscriber has to be passed on which the internal handler gets registered.
// A topic needs to be passed from which the subscriber will receive events.
// The method returns a function to unsubscribe the handler.
func On[T any](handlerFn func(event T) error, subscriber Subscriber, topic string) (UnsubscribeFn, error) {
	handler := func(e cloudevents.Event) error {
		var event T
		if err := e.DataAs(&event); err != nil {
			return nil
		}
		return handlerFn(event)
	}

	if err := subscriber.Subscribe(topic, handler); err != nil {
		return func() {}, err
	}

	unsubscribe := func() {
		_ = subscriber.Unsubscribe(handler)
	}

	return unsubscribe, nil
}

// Chan is a convenience function for creating a channel to an event handler.
// A subscriber has to be passed on which the internal handler gets registered.
// A topic needs to be passed from which the subscriber will receive events.
// A match function can be passed to filter events, reaching the channel.
// The method returns a channel and a function to unsubscribe the handler.
func Chan[T any](subscriber Subscriber, topic string, match ...func(event T) bool) (<-chan T, UnsubscribeFn, error) {
	eventChan := make(chan T)

	handler := func(e cloudevents.Event) error {
		var event T
		if err := e.DataAs(&event); err != nil {
			return nil
		}
		if len(match) == 1 && !match[0](event) {
			return nil
		}
		eventChan <- event
		return nil
	}

	if err := subscriber.Subscribe(topic, handler); err != nil {
		return nil, func() {
			close(eventChan)
		}, err
	}

	unsubscribe := func() {
		close(eventChan)
		_ = subscriber.Unsubscribe(handler)
	}

	return eventChan, unsubscribe, nil
}
