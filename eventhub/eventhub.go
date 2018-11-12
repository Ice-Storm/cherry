package eventhub

type Pubsub interface {
	Sub(topics ...string) chan interface{}
	SubOnce(topics ...string) chan interface{}
	Unsub(ch chan interface{}, topics ...string) error
	Pub(msg interface{}, topics ...string) error
	Close() error
	Shutdown() error
}

type EventHub struct {
	Pubsub
	Provider
}
