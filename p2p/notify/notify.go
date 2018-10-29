package notify

import (
	"sync"

	"cherrychain/p2p/eventhub"

	logging "cherrychain/common/clogging"

	inet "github.com/libp2p/go-libp2p-net"
	multiaddr "github.com/multiformats/go-multiaddr"
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
	SysEventMaxSize = 100
	ReadBuf         = 100
	WriteBuf        = 100
	SYS             = "SYSTEM"
	WRITE           = "WRITE"
	READ            = "READ"
	NetworkListen   = iota
	NetworkConnected
	NetworkDisconnected
	NetworkOpenedStream
	NetworkClosedStream
)

var (
	once sync.Once
	eh   Notify
	log  = logging.MustGetLogger("NOTIFY")
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

// SysListen binding p2p netwrok listen  event
func (n *Notify) SysListen(network inet.Network, ma multiaddr.Multiaddr) {
	n.Notifee.ListenF = func(inet.Network, multiaddr.Multiaddr) {
		log.Info("System listen event")
		n.SysEventHub.Pub(&SysEvent{
			SysType: NetworkListen,
			Meta:    ma,
		}, SYS)
	}
}

// SysConnected binding p2p netwrok Connected  event
func (n *Notify) SysConnected(network inet.Network, s inet.Stream) {
	n.Notifee.ConnectedF = func(inet.Network, inet.Conn) {
		log.Info("System Connected event")
		n.SysEventHub.Pub(&SysEvent{
			SysType: NetworkConnected,
			Meta:    s,
		}, SYS)
	}
}

// SysDisconnected binding p2p netwrok Disconnected  event
func (n *Notify) SysDisconnected(network inet.Network, s inet.Stream) {
	n.Notifee.DisconnectedF = func(inet.Network, inet.Conn) {
		log.Info("System Disconnected event")
		n.SysEventHub.Pub(&SysEvent{
			SysType: NetworkDisconnected,
			Meta:    s,
		}, SYS)
	}
}

// SysOpenedStream binding p2p netwrok OpenedStream  event
func (n *Notify) SysOpenedStream(network inet.Network, s inet.Stream) {
	n.Notifee.OpenedStreamF = func(inet.Network, inet.Stream) {
		log.Info("System OpenedStream event")
		n.SysEventHub.Pub(&SysEvent{
			SysType: NetworkOpenedStream,
			Meta:    s,
		}, SYS)
	}
}

// SysClosedStream binding p2p netwrok closeStream event
func (n *Notify) SysClosedStream(network inet.Network, s inet.Stream) {
	n.Notifee.ClosedStreamF = func(inet.Network, inet.Stream) {
		log.Info("System ClosedStream event")
		n.SysEventHub.Pub(&SysEvent{
			SysType: NetworkClosedStream,
			Meta:    s,
		}, SYS)
	}
}
