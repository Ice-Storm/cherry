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
func (e *Provider) Sub(topics ...string) chan interface{} {
	return e.PubSub.Sub(topics...)
}

// Pub publishes the given message to all subscribers of the specified topics.
func (e *Provider) Pub(msg interface{}, topics ...string) {
	e.PubSub.Pub(msg, topics...)
}

func (e *Provider) Unsub(ch chan interface{}, topics ...string) {
	e.PubSub.Unsub(ch, topics...)
}

// Close channel
func (e *Provider) Close(topics ...string) {
	e.PubSub.Close(topics...)
}

// Shutdown closes all channels
func (e *Provider) Shutdown() {
	e.PubSub.Shutdown()
}
