package eventhub

import observer "github.com/cskr/pubsub"

type Provider struct {
	PubSub *observer.PubSub
}

// New creates a new PubSub
func New(capacity int) *Provider {
	return &Provider{PubSub: observer.New(capacity)}
}

// Sub returns a channel on which messages published on any of the specified topics.
func (e *Provider) Sub(topics ...string) (chan interface{}, error) {
	return e.PubSub.Sub(topics...), nil
}

// Pub publishes the given message to all subscribers of the specified topics.
func (e *Provider) Pub(msg interface{}, topics ...string) error {
	e.PubSub.Pub(msg, topics...)
	return nil
}

func (e *Provider) Unsub(ch chan interface{}, topics ...string) error {
	e.PubSub.Unsub(ch, topics...)
	return nil
}

// Close channel
func (e *Provider) Close(topics ...string) error {
	e.PubSub.Close(topics...)
	return nil
}

// Shutdown closes all channels
func (e *Provider) Shutdown() error {
	e.PubSub.Shutdown()
	return nil
}
