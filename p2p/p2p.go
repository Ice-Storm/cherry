package p2p

import (
	"context"
	"crypto/rand"
	"fmt"
	"sync"

	"cherrychain/common/clogging"
	"cherrychain/p2p/notify"

	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-host"
	inet "github.com/libp2p/go-libp2p-net"
	multiaddr "github.com/multiformats/go-multiaddr"
)

var p2pLogger = clogging.MustGetLogger("P2P")

var mutex = &sync.Mutex{}

const MessageSizeMax = 1 << 22 // 4 MB

type P2P struct {
	Host   host.Host
	Notify *notify.Notify
}

func New(ctx context.Context, genesisMultiAddr string) *P2P {
	p2pLogger.Debug("New p2p module")
	p2pLogger.Info("Cherrychain start .....")

	sourceMultiAddr, err := multiaddr.NewMultiaddr(genesisMultiAddr)

	if err != nil {
		p2pLogger.Fatal("Invalid address: ", err)
	}

	host, err := genesisNode(ctx, sourceMultiAddr)

	if err != nil {
		p2pLogger.Fatal("Cant't create p2p module: ", err)
	}

	nt, err := notify.New()

	if err != nil {
		p2pLogger.Fatal("Cant't create p2p notify module: ", err)
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
		p2pLogger.Error("Cant't generate node private key")
	}

	host, _ := libp2p.New(
		ctx,
		libp2p.ListenAddrs(genesisMultiAddr),
		libp2p.Identity(prvKey),
	)
	return host, nil
}

func (n *P2P) HandleStream(s inet.Stream) {
	p2pLogger.Info("Open new stream")
	n.Notify.SysOpenedStream(n.Host.Network(), s)
	n.Notify.Notifee.OpenedStream(n.Host.Network(), s)
}

func (n *P2P) StartSysEventLoop() {
	sysEvent, _ := n.Notify.SysEventHub.Sub(notify.SYS)
	go func() {
		for event := range sysEvent {
			switch (event.(*notify.SysEvent)).SysType {
			case notify.NetworkConnected:
				fmt.Printf("NetworkConnected ***")
			case notify.NetworkOpenedStream:
				n.broadcast(((event.(*notify.SysEvent)).Meta).(inet.Stream))
			default:
				panic("Invalid system event type")
			}
		}
	}()
}

func (n *P2P) broadcast(s inet.Stream) {
	go func(s inet.Stream) {
		defer s.Close()
		msgChan, _ := n.Notify.WritePB.Sub(notify.WRITE)
		n.readData(s)
		for msg := range msgChan {
			p2pLogger.Debug("p2p network broadcast message", msg.([]byte))
			s.Write(msg.([]byte))
		}
	}(s)
}

func (n *P2P) readData(s inet.Stream) {
	go func() {
		msg := make([]byte, MessageSizeMax)
		for {
			rn, err := s.Read(msg)
			if err != nil || rn == 0 {
				p2pLogger.Debug("Read error", err)
				return
			}
			n.Notify.ReadPB.Pub(msg, notify.READ)
		}
	}()
}

func (n *P2P) Write(data []byte) {
	n.Notify.WritePB.Pub(data, notify.WRITE)
}

func (n *P2P) Read(cap []byte) (int, error) {
	msgChan, err := n.Notify.ReadPB.Sub(notify.READ)
	if err != nil {
		return 0, err
	}
	for msg := range msgChan {
		msgByte := msg.([]byte)
		if len(msgByte) == 0 {
			continue
		}
		return copy(cap, msgByte), nil
	}
	return 0, nil
}
