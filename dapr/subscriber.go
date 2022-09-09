package dapr

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/Becklyn/eventor"
	"github.com/Becklyn/eventor/tracing"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

var (
	ErrSubscriberNotFound = fmt.Errorf("subscriber not found")
)

type daprSubscriber struct {
	Pubsub   string              `json:"pubsubname"`
	Topic    string              `json:"topic"`
	Route    string              `json:"route"`
	Handlers []eventor.HandlerFn `json:"-"`
}

// DaprSubscriberRegistryConfig is the configuration struct for a subscriber registry.
type DaprSubscriberRegistryConfig struct {
	App    *fiber.App
	Pubsub string
}

type daprSubscriberRegistry struct {
	mux         sync.Mutex
	subscribers []*daprSubscriber
	app         *fiber.App
	pubsub      string
}

// NewDaprSubscriberRegistry creates a new subscriber registry instance.
func NewDaprSubscriberRegistry(config DaprSubscriberRegistryConfig) *daprSubscriberRegistry {
	subscriberRegistry := &daprSubscriberRegistry{
		subscribers: []*daprSubscriber{},
		app:         config.App,
		pubsub:      config.Pubsub,
	}

	subscriberRegistry.app.Get("/dapr/subscribe", func(ctx *fiber.Ctx) error {
		return ctx.JSON(subscriberRegistry.subscribers)
	})

	return subscriberRegistry
}

func (r *daprSubscriberRegistry) Subscribe(topic string, handler eventor.HandlerFn) error {
	r.mux.Lock()
	defer r.mux.Unlock()

	subscriberForTopic, subscriberForTopicFound := lo.Find(r.subscribers, func(s *daprSubscriber) bool {
		return s.Topic == topic
	})
	if !subscriberForTopicFound {
		newSubscriber := &daprSubscriber{
			Pubsub:   r.pubsub,
			Topic:    topic,
			Route:    fmt.Sprintf("/dapr/%s/%s", r.pubsub, topic),
			Handlers: []eventor.HandlerFn{handler},
		}

		r.bind(newSubscriber)
		r.subscribers = append(r.subscribers, newSubscriber)

		return nil
	}

	subscriberForTopic.Handlers = append(subscriberForTopic.Handlers, handler)
	r.bind(subscriberForTopic)

	return nil
}

func (r *daprSubscriberRegistry) Unsubscribe(handler eventor.HandlerFn) error {
	r.mux.Lock()
	defer r.mux.Unlock()

	for _, subscriber := range r.subscribers {
		for i, h := range subscriber.Handlers {
			refH := reflect.ValueOf(h)
			refHandler := reflect.ValueOf(handler)
			if refH.Pointer() != refHandler.Pointer() {
				continue
			}

			subscriber.Handlers = append(subscriber.Handlers[:i], subscriber.Handlers[i+1:]...)
			r.bind(subscriber)

			return nil
		}
	}

	return ErrSubscriberNotFound
}

func (r *daprSubscriberRegistry) bind(subscriber *daprSubscriber) {
	r.app.Post(subscriber.Route, func(ctx *fiber.Ctx) error {
		event := cloudevents.NewEvent()
		if err := event.UnmarshalJSON(ctx.Body()); err != nil {
			return ctx.SendStatus(fiber.StatusInternalServerError)
		}

		spanCtx := tracing.FromFiberContext(ctx)

		wg := sync.WaitGroup{}
		errors := []error{}

		for _, handler := range subscriber.Handlers {
			wg.Add(1)
			go func(handler eventor.HandlerFn) {
				defer wg.Done()
				if err := handler(event, spanCtx); err != nil {
					errors = append(errors, err)
				}
			}(handler)
		}

		wg.Wait()

		if len(errors) > 0 {
			return ctx.SendStatus(fiber.StatusBadRequest)
		}
		return ctx.SendStatus(fiber.StatusOK)
	}).Name(subscriber.Topic)
}
