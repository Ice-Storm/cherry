package p2p

import (
	"bufio"
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"sync"
	"time"

	"cherrychain/common/clogging"
	"cherrychain/p2p/notify"

	"github.com/juju/ratelimit"
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
	Host      host.Host
	RateLimit *ratelimit.Bucket
	Notify    *notify.Notify
}

func New(ctx context.Context, genesisMultiAddr string) *P2P {
	p2pLogger.Debug("New p2p module")

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

	nt.Notifee.ListenF = func(inet.Network, multiaddr.Multiaddr) {
		p2pLogger.Info("Cherrychain start .....")
	}

	nt.Notifee.Listen(host.Network(), sourceMultiAddr)

	return &P2P{
		Host:      host,
		RateLimit: ratelimit.NewBucketWithRate(5, int64(100)),
		Notify:    nt,
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
	sysEvent, _ := n.Notify.SysEventHub.Sub(notify.SYS_CHAN_TYPE)
	n.Notify.PubSysOpenedStream(n.Host.Network(), s)
	n.Notify.Notifee.OpenedStream(n.Host.Network(), s)
	go func() {
		for event := range sysEvent {
			switch (event.(*notify.SysEvent)).SysType {
			case notify.NetworkOpenedStream:
				n.broadcast(((event.(*notify.SysEvent)).Meta).(inet.Stream))
			default:
				panic("Invalid system event type")
			}
		}
	}()
}

func (n *P2P) broadcast(s inet.Stream) {
	msgChan, _ := n.Notify.UserEventHub.Sub(notify.USER_CHAN_TYPE)
	n.swapPeersInfo(s)

	for msg := range msgChan {
		p2pLogger.Debug("p2p network broadcast message", msg.([]byte))
		s.Write(msg.([]byte))
	}
}

func (n *P2P) swapPeersInfo(s inet.Stream) {
	n.readData(s)
	n.writeData(s)
}

func (n *P2P) readData(s inet.Stream) {
	go func() {
		bb := make([]byte, MessageSizeMax)

		for {
			if _, isTake := n.RateLimit.TakeMaxDuration(1, 500*time.Millisecond); !isTake {
				continue
			}
			n, err := s.Read(bb)
			if err != nil {
				p2pLogger.Error("Read error", err)
				return
			}
			if n == 0 {
				return
			}
			fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bb))
		}
	}()
}

func (n *P2P) writeData(s inet.Stream) {
	stdReader := bufio.NewReader(os.Stdin)
	go func() {
		for {
			fmt.Print("> ")
			sendData, err := stdReader.ReadString('\n')
			if err != nil {
				panic(err)
			}
			n.Notify.UserEventHub.Pub([]byte(sendData), notify.USER_CHAN_TYPE)
		}
	}()
}
