package dapr

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/Becklyn/eventor/v2/tracing"
	"go.opentelemetry.io/otel/trace"
)

var (
	ErrMessageForbidden     = errors.New("message forbidden by access controls")
	ErrNoPubsubOrTopic      = errors.New("no pubsub name or topic given")
	ErrDeliveryFailed       = errors.New("delivery failed")
	ErrUnexpectedStatusCode = errors.New("unexpected status code")
)

// DaprPublisherConfig is the configuration struct for a publisher.
type DaprPublisherConfig struct {
	Host   string
	Pubsub string
}

type daprPublisher struct {
	client *http.Client
	host   string
	pubsub string
}

// NewDaprPublisher creates a new publisher instance.
func NewDaprPublisher(config DaprPublisherConfig) *daprPublisher {
	return &daprPublisher{
		client: http.DefaultClient,
		host:   config.Host,
		pubsub: config.Pubsub,
	}
}

func (p *daprPublisher) Publish(topic string, data any, spanCtx ...trace.SpanContext) error {
	url := fmt.Sprintf(
		"%s/v1.0/publish/%s/%s",
		p.host,
		p.pubsub,
		topic,
	)

	body, err := json.Marshal(data)
	if err != nil {
		return err
	}
	bodyAsBuffer := bytes.NewBuffer(body)

	req, err := http.NewRequest(
		"POST",
		url,
		bodyAsBuffer,
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	if len(spanCtx) == 1 && spanCtx[0].IsValid() {
		traceparent, treacestate := tracing.ToHeaders(spanCtx[0])
		req.Header.Set("traceparent", traceparent)
		if len(treacestate) > 0 {
			req.Header.Set("tracestate", treacestate)
		}
	}

	res, err := p.client.Do(req)
	if err != nil {
		return err
	}

	switch res.StatusCode {
	case 403:
		return ErrMessageForbidden
	case 404:
		return ErrNoPubsubOrTopic
	case 500:
		return ErrDeliveryFailed
	case 204:
		return nil // Message delivered
	default:
		return ErrUnexpectedStatusCode
	}
}
