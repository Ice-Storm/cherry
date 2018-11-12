package eventhub

type Pubsub interface {
	Sub(topics ...string) chan interface{}
	SubOnce(topics ...string) chan interface{}
	Unsub(ch chan interface{}, topics ...string)
	Pub(msg interface{}, topics ...string)
	Close()
	Shutdown()
}

type EventHub struct {
	Pubsub
	Provider
}
