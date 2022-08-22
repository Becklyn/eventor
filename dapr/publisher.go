package dapr

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrMessageForbidden     = errors.New("message forbidden by access controls")
	ErrNoPubsubOrTopic      = errors.New("no pubsub name or topic given")
	ErrDeliveryFailed       = errors.New("delivery failed")
	ErrUnexpectedStatusCode = errors.New("unexpected status code")
)

// DaprPublisherConfig is the configuration struct for a publisher.
type DaprPublisherConfig struct {
	Hostname string
	Port     string
	Pubsub   string
}

type daprPublisher struct {
	client   *http.Client
	hostname string
	port     string
	pubsub   string
}

// NewDaprPublisher creates a new publisher instance.
func NewDaprPublisher(config DaprPublisherConfig) *daprPublisher {
	return &daprPublisher{
		client:   http.DefaultClient,
		hostname: config.Hostname,
		port:     config.Port,
		pubsub:   config.Pubsub,
	}
}

func (p *daprPublisher) Publish(topic string, data any) error {
	url := fmt.Sprintf(
		"http://%s:%s/v1.0/publish/%s/%s",
		p.hostname,
		p.port,
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
