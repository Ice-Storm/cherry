package notify

import (
	"sync"

	"cherrychain/common/clogging"
	"cherrychain/p2p/eventhub"

	inet "github.com/libp2p/go-libp2p-net"
)

type Notify struct {
	Notifee     inet.NotifyBundle
	SysEventHub *eventhub.Provider
	WritePB     *eventhub.Provider
	ReadPB      *eventhub.Provider
}

type SysEvent struct {
	SysType int
	Meta    interface{}
}

const (
	SysEventMaxSize     = 100
	ReadBuf             = 100
	WriteBuf            = 100
	SYS                 = "SYSTEM"
	WRITE               = "WRITE"
	READ                = "READ"
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
			Notifee:     notifee,
			SysEventHub: eventhub.New(SysEventMaxSize),
			WritePB:     eventhub.New(WriteBuf),
			ReadPB:      eventhub.New(ReadBuf),
		}
	})
	return &eh, nil
}

func (n *Notify) SysOpenedStream(network inet.Network, s inet.Stream) {
	n.Notifee.OpenedStreamF = func(inet.Network, inet.Stream) {
		notifyLogger.Info("OpenedStream....")
		n.SysEventHub.Pub(&SysEvent{
			SysType: NetworkOpenedStream,
			Meta:    s,
		}, SYS)
	}
}
