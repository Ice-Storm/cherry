package notify

import (
	"sync"

	inet "github.com/libp2p/go-libp2p-net"
)

type Notify struct {
	Notifee inet.NotifyBundle
}

var once sync.Once
var eh Notify

func New() (*Notify, error) {
	once.Do(func() {
		var notifee inet.NotifyBundle
		eh = Notify{
			Notifee: notifee,
		}
	})
	return &eh, nil
}
