package p2p

import (
	"context"
	"crypto/rand"
	"errors"
	"sync"

	"cherrychain/p2p/notify"

	logging "cherrychain/common/clogging"

	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-host"
	inet "github.com/libp2p/go-libp2p-net"
	multiaddr "github.com/multiformats/go-multiaddr"
)

var (
	mutex = &sync.Mutex{}
	log   = logging.MustGetLogger("P2P")
)

const MessageSizeMax = 1 << 22 // 4 MB

type P2P struct {
	Host   host.Host
	Notify *notify.Notify
}

func New(ctx context.Context, genesisMultiAddr string) *P2P {
	log.Info("Cherrychain start .....")

	sourceMultiAddr, err := multiaddr.NewMultiaddr(genesisMultiAddr)

	if err != nil {
		log.Fatal("Invalid address: ", err)
	}

	host, err := genesisNode(ctx, sourceMultiAddr)

	if err != nil {
		log.Fatal("Cant't create p2p module: ", err)
	}

	nt, err := notify.New()

	if err != nil {
		log.Fatal("Cant't create p2p notify module: ", err)
	}

	// Bind system event buf
	nt.SysListen(host.Network(), sourceMultiAddr)

	// Emit system listen event
	nt.Notifee.Listen(host.Network(), sourceMultiAddr)

	return &P2P{
		Host:   host,
		Notify: nt,
	}
}

func genesisNode(ctx context.Context, genesisMultiAddr multiaddr.Multiaddr) (host.Host, error) {
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)

	if err != nil {
		log.Error("Cant't generate node private key")
	}

	host, _ := libp2p.New(
		ctx,
		libp2p.ListenAddrs(genesisMultiAddr),
		libp2p.Identity(prvKey),
	)
	return host, nil
}

func (n *P2P) HandleStream(s inet.Stream) {
	log.Info("Open new stream")
	n.Notify.SysOpenedStream(n.Host.Network(), s)
	n.Notify.Notifee.OpenedStream(n.Host.Network(), s)
}

func (n *P2P) StartSysEventLoop(ctx context.Context) error {
	sysEvent, err := n.Notify.SysEventHub.Sub(notify.SYS)
	if err != nil {
		return errors.New("Cant not subscribe queue")
	}
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
	switch (event.(*notify.SysEvent)).SysType {
	case notify.NetworkConnected:
		n.Notify.SysDisconnected(n.Host.Network(), ((event.(*notify.SysEvent)).Meta).(inet.Stream))
	case notify.NetworkDisconnected:
		n.closeConnection(((event.(*notify.SysEvent)).Meta).(inet.Stream))
	case notify.NetworkOpenedStream:
		n.Notify.SysClosedStream(n.Host.Network(), ((event.(*notify.SysEvent)).Meta).(inet.Stream))
		n.broadcast(((event.(*notify.SysEvent)).Meta).(inet.Stream))
	case notify.NetworkClosedStream:
		n.closeStream(((event.(*notify.SysEvent)).Meta).(inet.Stream))
	default:
		panic("Invalid system event type")
	}
}

func (n *P2P) broadcast(s inet.Stream) {
	go func(s inet.Stream) {
		defer s.Close()
		msgChan, err := n.Notify.WritePB.Sub(notify.WRITE)
		if err != nil {
			return
		}
		n.readData(s)
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

func (n *P2P) Write(data []byte) error {
	return n.Notify.WritePB.Pub(data, notify.WRITE)
}

func (n *P2P) Read(cap []byte) (int, error) {
	msgChan, err := n.Notify.ReadPB.Sub(notify.READ)
	if err != nil {
		return 0, err
	}
	for msg := range msgChan {
		msgBytes := msg.([]byte)
		if len(msgBytes) == 0 {
			continue
		}
		return copy(cap, msgBytes), nil
	}
	return 0, nil
}

func (n *P2P) CloseStream(s inet.Stream) error {
	n.Notify.Notifee.ClosedStream(n.Host.Network(), s)
	return nil
}

func (n *P2P) CloseConnection(s inet.Stream) error {
	n.Notify.Notifee.Disconnected(n.Host.Network(), s.Conn())
	return nil
}
