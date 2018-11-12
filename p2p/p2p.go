package p2p

import (
	"context"
	"crypto/rand"
	"sync"

	"cherrychain/notify"

	logging "cherrychain/clogging"

	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-host"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	inet "github.com/libp2p/go-libp2p-net"
	multiaddr "github.com/multiformats/go-multiaddr"
)

var mutex = &sync.Mutex{}
var log = logging.MustGetLogger("P2P")

const MessageSizeMax = 1 << 22 // 4 MB

// P2P network
type P2P struct {
	Host   host.Host
	Notify *notify.Notify
}

// New creater p2p network
func New(ctx context.Context, addr string, isGenesisNode bool) *P2P {
	log.Info("Cherrychain start .....")

	sourceMultiAddr, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		log.Fatal("Invalid address: ", err)
	}

	host, err := newNode(ctx, sourceMultiAddr)
	if err != nil {
		log.Fatal("Cant't create p2p module: ", err)
	}

	nt, err := notify.New()
	if err != nil {
		log.Fatal("Cant't create p2p notify module: ", err)
	}

	// Binding system event buf
	nt.SysListen(host.Network(), sourceMultiAddr)

	// Emit system listen event
	nt.Notifee.Listen(host.Network(), sourceMultiAddr)

	if isGenesisNode {
		if _, err := dht.New(ctx, host); err != nil {
			log.Fatal("Can not create genesis node: ", err)
		}
		log.Info("Create genesis Node. ")
	}

	return &P2P{
		Host:   host,
		Notify: nt,
	}
}

func newNode(ctx context.Context, addr multiaddr.Multiaddr) (host.Host, error) {
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
	if err != nil {
		log.Error("Cant't generate node private key")
	}

	return libp2p.New(
		ctx,
		libp2p.ListenAddrs(addr),
		libp2p.Identity(prvKey),
	)
}

// HandleStream handle new stream incoming to other peer
func (n *P2P) HandleStream(s inet.Stream) {
	log.Info("Open new stream")
	n.Notify.SysOpenedStream(n.Host.Network(), s)
	n.Notify.Notifee.OpenedStream(n.Host.Network(), s)
}

// StartSysEventLoop deal with system event. eg network connected
func (n *P2P) StartSysEventLoop(ctx context.Context) error {
	sysEvent := n.Notify.SysEventHub.Sub(notify.SYS)
	go func() {
		for {
			select {
			case event := <-sysEvent:
				n.eventDestribute(event)
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}

func (n *P2P) eventDestribute(event interface{}) {
	stream := ((event.(*notify.SysEvent)).Meta).(inet.Stream)
	switch (event.(*notify.SysEvent)).SysType {
	case notify.NetworkConnected:
		n.Notify.SysDisconnected(n.Host.Network(), stream)
	case notify.NetworkDisconnected:
		n.closeConnection(stream)
	case notify.NetworkOpenedStream:
		n.Notify.SysClosedStream(n.Host.Network(), stream)
		n.broadcast(stream)
	case notify.NetworkClosedStream:
		n.closeStream(stream)
	default:
		panic("Invalid system event type")
	}
}

func (n *P2P) broadcast(s inet.Stream) {
	go func(s inet.Stream) {
		defer s.Close()
		n.readData(s)
		msgChan := n.Notify.WritePB.Sub(notify.WRITE)
		for msg := range msgChan {
			s.Write(msg.([]byte))
		}
	}(s)
}

func (n *P2P) closeStream(s inet.Stream) {
	s.Close()
}

func (n *P2P) closeConnection(s inet.Stream) {
	s.Conn().Close()
}

func (n *P2P) readData(s inet.Stream) {
	go func() {
		msg := make([]byte, MessageSizeMax)
		for {
			rn, err := s.Read(msg)
			if err != nil || rn == 0 {
				log.Error("Read error", err)
				return
			}
			n.Notify.ReadPB.Pub(msg, notify.READ)
		}
	}()
}

// Write is used to send message to other what is be connedted to this peer
func (n *P2P) Write(data []byte) {
	n.Notify.WritePB.Pub(data, notify.WRITE)
}

// Read is used to receive message to other what is be connedted to this peer
func (n *P2P) Read(cap []byte) (int, error) {
	msgChan := n.Notify.ReadPB.Sub(notify.READ)
	defer n.Notify.ReadPB.Unsub(msgChan, notify.READ)
	for msg := range msgChan {
		msgBytes := msg.([]byte)
		if len(msgBytes) == 0 {
			continue
		}
		return copy(cap, msgBytes), nil
	}
	return 0, nil
}

// CloseStream close stream
func (n *P2P) CloseStream(s inet.Stream) error {
	n.Notify.Notifee.ClosedStream(n.Host.Network(), s)
	return nil
}

// CloseConnection close connection
func (n *P2P) CloseConnection(s inet.Stream) error {
	n.Notify.Notifee.Disconnected(n.Host.Network(), s.Conn())
	return nil
}
