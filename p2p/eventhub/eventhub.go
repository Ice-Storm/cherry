package eventhub

type Pubsub interface {
	Sub(topics ...string) chan interface{}
	SubOnce(topics ...string) chan interface{}
	Unsub(topics ...string) error
	Pub(msg interface{}, topics ...string) error
	Close() error
}

type EventHub struct {
	Pubsub
	Provider
}
