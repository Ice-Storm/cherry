package eventhub

import observer "github.com/cskr/pubsub"

type Provider struct {
	PubSub *observer.PubSub
}

func New(capacity int) *Provider {
	return &Provider{PubSub: observer.New(capacity)}
}

func (e *Provider) Sub(topics ...string) (chan interface{}, error) {
	return e.PubSub.Sub(topics...), nil
}

func (e *Provider) Pub(msg interface{}, topics ...string) error {
	e.PubSub.Pub(msg, topics...)
	return nil
}
