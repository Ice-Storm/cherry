package p2p

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"cherrychain/common/clogging"

	"github.com/juju/ratelimit"
	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-host"
	libnet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	multiaddr "github.com/multiformats/go-multiaddr"
)

var p2pLogger = clogging.MustGetLogger("P2P")

var mutex = &sync.Mutex{}

const protocolID = "/cherryCahin/1.0"

type P2P struct {
	Host      host.Host
	PeerStore *PeerStore
	RateLimit *ratelimit.Bucket
}

func New(genesisMultiAddr string) *P2P {
	p2pLogger.Info("New p2p module")
	host, err := genesisNode(genesisMultiAddr)

	if err != nil {
		p2pLogger.Error("Cant't new p2p module")
	}

	return &P2P{
		Host:      host,
		PeerStore: NewPeerStore(),
		RateLimit: ratelimit.NewBucketWithRate(5, int64(100)),
	}
}

func genesisNode(genesisMultiAddr string) (host.Host, error) {
	sourceMultiAddr, _ := multiaddr.NewMultiaddr(genesisMultiAddr)

	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)

	if err != nil {
		p2pLogger.Error("Cant't generate node private key")
	}

	host, _ := libp2p.New(
		context.Background(),
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
	)
	return host, nil
}

func (p *P2P) HandleStream(s libnet.Stream) {
	p.swapPeersInfo(s)
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

func (n *P2P) IsNetworkMetaData(bytes []byte) error {
	validateLen := 1024 * 1024

	if len(bytes) == 0 {
		return errors.New("Empty message")
	}

	if len(bytes) > validateLen {
		p2pLogger.Error("Beyond max size")
		return errors.New("Beyond max size")
	}

	return n.verifyPeerDiscovery(bytes)
}

func (n *P2P) verifyPeerDiscovery(bytes []byte) error {
	var peerInfo PeerDiscovery
	if err := json.Unmarshal(bytes, &peerInfo); err != nil {
		p2pLogger.Debug("P2P network message is invalidete")
		return err
	}

	addr := fmt.Sprintf("/ip4/%s/tcp/%d/ipfs/%s", peerInfo.IP, peerInfo.Port, peerInfo.ID)

	cherryAddr, err := multiaddr.NewMultiaddr(addr)

	if err != nil {
		p2pLogger.Debug("P2P network message is invalidete")
		return err
	}

	if _, err := cherryAddr.ValueForProtocol(multiaddr.P_IPFS); err != nil {
		p2pLogger.Debug("Message is not peerDiscovery type")
		return err
	}

	return nil
}

func (n *P2P) AddAddrToPeerstore(h host.Host, addr string) peer.ID {
	cherryAddr, err := multiaddr.NewMultiaddr(addr)

	if err != nil {
		p2pLogger.Error(err)
	}

	pid, err := cherryAddr.ValueForProtocol(multiaddr.P_IPFS)

	if err != nil {
		p2pLogger.Error(err)
	}

	peerid, err := peer.IDB58Decode(pid)

	if err != nil {
		p2pLogger.Error(err)
	}

	targetPeerAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ipfs/%s", peer.IDB58Encode(peerid)))
	targetAddr := cherryAddr.Decapsulate(targetPeerAddr)

	h.Peerstore().AddAddr(peerid, targetAddr, peerstore.PermanentAddrTTL)
	return peerid
}

func (n *P2P) swapPeersInfo(s libnet.Stream) {
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

			if err := n.IsNetworkMetaData([]byte(str)); err != nil {
				p2pLogger.Debug("Invalidete network message")
				return
			}

			var peerInfo PeerDiscovery
			json.Unmarshal([]byte(str), &peerInfo)

			if _, err := n.PeerStore.Push(peerInfo); err != nil {
				p2pLogger.Info("Failed push new peer info")
			} else {
				fmt.Printf("\nPeerInfo %v \n", peerInfo)
				fmt.Printf("Host %s \n", n.Host.ID().Pretty())

				if peerInfo.ID != n.Host.ID().Pretty() {
					fmt.Printf("\n/ip4/%s/tcp/%d/ipfs/%s\n", peerInfo.IP, peerInfo.Port, peerInfo.ID)
					n.attach(peerInfo.IP, peerInfo.Port, peerInfo.ID)
				}
			}
			fmt.Printf("\nlen : %d\n", len(n.PeerStore.Store))
		}
	}()
}

func (n *P2P) writeData(s libnet.Stream) {
	rw := bufio.NewWriter(s)
	go func() {
		for {
			for i := 0; i < len(n.PeerStore.Store); i++ {
				time.Sleep(5 * time.Second)
				mutex.Lock()
				data, err := json.Marshal(n.PeerStore.Store[i])
				if err != nil {
					p2pLogger.Error("Failed marshal peer store")
				}
				n.WriteBytes(rw, data)
				mutex.Unlock()
				if i > len(n.PeerStore.Store) {
					i = 0
				}
			}
		}
	}()
}

func (n *P2P) attach(ip string, port uint16, hostID string) error {
	peerID := n.AddAddrToPeerstore(n.Host, fmt.Sprintf("/ip4/%s/tcp/%d/ipfs/%s", ip, port, hostID))
	s, err := n.Host.NewStream(context.Background(), peerID, protocolID)
	if err != nil {
		panic(err)
	}
	n.swapPeersInfo(s)
	return err
}
