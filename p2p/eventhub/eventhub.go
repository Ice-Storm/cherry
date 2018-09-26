package eventhub

import (
	"sync"

	inet "github.com/libp2p/go-libp2p-net"
)

type EventHub struct {
	Notifee inet.NotifyBundle
}

var once sync.Once
var eh EventHub

func New() (*EventHub, error) {
	once.Do(func() {
		var notifee inet.NotifyBundle
		eh = EventHub{
			Notifee: notifee,
		}
	})
	return &eh, nil
}
