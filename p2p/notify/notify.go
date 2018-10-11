package notify

import (
	"cherrychain/p2p/eventhub"
	"fmt"
	"sync"

	inet "github.com/libp2p/go-libp2p-net"
)

type Notify struct {
	Notifee      inet.NotifyBundle
	SysEventHub  *eventhub.Provider
	UserEventHub *eventhub.Provider
}

type SysEvent struct {
	SysType int
	Meta    interface{}
}

const (
	NetworkOpenedStream = iota
)

var (
	once sync.Once
	eh   Notify
)

func New() (*Notify, error) {
	once.Do(func() {
		var notifee inet.NotifyBundle
		eh = Notify{
			Notifee:      notifee,
			SysEventHub:  eventhub.New(100),
			UserEventHub: eventhub.New(100),
		}
	})
	return &eh, nil
}

func (n *Notify) PubSysOpenedStream(network inet.Network, s inet.Stream) {
	n.Notifee.OpenedStreamF = func(inet.Network, inet.Stream) {
		fmt.Printf("PubOpenedStream....")
		n.SysEventHub.Pub(&SysEvent{
			SysType: NetworkOpenedStream,
			Meta:    s,
		}, "system")
	}
}
