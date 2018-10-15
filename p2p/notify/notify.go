package notify

import (
	"sync"

	"cherrychain/common/clogging"
	"cherrychain/p2p/eventhub"

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
	SysEventMaxSize     = 100
	UserEventMaxSize    = 100
	SYS_CHAN_TYPE       = "SYSTEM"
	USER_CHAN_TYPE      = "USER"
	NetworkOpenedStream = iota
)

var (
	once         sync.Once
	eh           Notify
	notifyLogger = clogging.MustGetLogger("NOTIFY")
)

func New() (*Notify, error) {
	once.Do(func() {
		var notifee inet.NotifyBundle
		eh = Notify{
			Notifee:      notifee,
			SysEventHub:  eventhub.New(SysEventMaxSize),
			UserEventHub: eventhub.New(UserEventMaxSize),
		}
	})
	return &eh, nil
}

func (n *Notify) PubSysOpenedStream(network inet.Network, s inet.Stream) {
	n.Notifee.OpenedStreamF = func(inet.Network, inet.Stream) {
		notifyLogger.Info("PubOpenedStream....")
		n.SysEventHub.Pub(&SysEvent{
			SysType: NetworkOpenedStream,
			Meta:    s,
		}, SYS_CHAN_TYPE)
	}
}
