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

	"github.com/juju/ratelimit"
	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-host"
	libnet "github.com/libp2p/go-libp2p-net"
	multiaddr "github.com/multiformats/go-multiaddr"
)

var p2pLogger = clogging.MustGetLogger("P2P")

var mutex = &sync.Mutex{}

const protocolID = "/cherryCahin/1.0"

type P2P struct {
	Host      host.Host
	RateLimit *ratelimit.Bucket
}

func New(ctx context.Context, genesisMultiAddr string) *P2P {
	p2pLogger.Info("New p2p module")
	host, err := genesisNode(ctx, genesisMultiAddr)

	if err != nil {
		p2pLogger.Error("Cant't new p2p module: ", err)
	}

	return &P2P{
		Host:      host,
		RateLimit: ratelimit.NewBucketWithRate(5, int64(100)),
	}
}

func genesisNode(ctx context.Context, genesisMultiAddr string) (host.Host, error) {
	sourceMultiAddr, _ := multiaddr.NewMultiaddr(genesisMultiAddr)

	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)

	if err != nil {
		p2pLogger.Error("Cant't generate node private key")
	}

	host, _ := libp2p.New(
		ctx,
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
	)
	return host, nil
}

func (n *P2P) HandleStream(s libnet.Stream) {
	n.swapPeersInfo(s)
}

func (n *P2P) ReadString(IO interface{}) (string, error) {
	switch stream := IO.(type) {
	case *bufio.ReadWriter:
		return stream.ReadString('\n')
	case *bufio.Reader:
		return stream.ReadString('\n')
	default:
		panic("Invalid IO interface")
	}
}

func (n *P2P) WriteBytes(stream *bufio.Writer, str []byte) error {
	n.WriteString(stream, string(append(str, '\n')))
	return nil
}

func (n *P2P) WriteString(stream *bufio.Writer, str string) error {
	stream.WriteString(str)
	return stream.Flush()
}

func (n *P2P) swapPeersInfo(s libnet.Stream) {
	p2pLogger.Info("Got a new stream!")
	n.readData(s)
	n.writeData(s)
}

func (n *P2P) readData(s libnet.Stream) {
	rw := bufio.NewReader(s)
	go func() {
		for {
			if _, isTake := n.RateLimit.TakeMaxDuration(1, 500*time.Millisecond); !isTake {
				continue
			}
			str, _ := n.ReadString(rw)
			if str == "" {
				return
			}
			if str != "\n" {
				fmt.Printf("\x1b[32m%s\x1b[0m> ", str)
			}
		}
	}()
}

func (n *P2P) writeData(s libnet.Stream) {
	rw := bufio.NewWriter(s)
	stdReader := bufio.NewReader(os.Stdin)
	go func() {
		for {
			fmt.Print("> ")
			sendData, err := stdReader.ReadString('\n')
			if err != nil {
				panic(err)
			}
			n.WriteBytes(rw, []byte(sendData))
		}
	}()
}
