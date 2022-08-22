package eventor

type Publisher interface {
	// Publish publishes any data to the given topic.
	// A topic needs to be passed to which the publisher publisheds the data.
	// The data can be of any arbitrary type.
	Publish(topic string, data any) error
}
