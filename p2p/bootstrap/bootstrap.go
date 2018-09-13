package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"cherrychain/common/clogging"
	"cherrychain/p2p"

	cid "github.com/ipfs/go-cid"
	ipfsaddr "github.com/ipfs/go-ipfs-addr"
	host "github.com/libp2p/go-libp2p-host"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
	multihash "github.com/multiformats/go-multihash"
)

var bootstrapLogger = clogging.MustGetLogger("bootstrap")
var preFix = "meet me here test"

type Config struct {
	BootstrapPeers    []string
	MinPeers          int
	BootstrapInterval time.Duration
}

type Bootstrap struct {
	minPeers       int
	bootstrapPeers []*peerstore.PeerInfo
	host           host.Host
}

// Create new bootstrap service
func New(h host.Host, c Config) (*Bootstrap, error) {

	if c.MinPeers > len(c.BootstrapPeers) {
		return nil, errors.New(fmt.Sprintf("Too less bootstrapping nodes. Expected at least: %d, got: %d", c.MinPeers, len(c.BootstrapPeers)))
	}

	var peers []*peerstore.PeerInfo

	for _, v := range c.BootstrapPeers {
		addr, err := ma.NewMultiaddr(v)

		if err != nil {
			return nil, err
		}

		pInfo, err := peerstore.InfoFromP2pAddr(addr)

		if err != nil {
			return nil, err
		}

		peers = append(peers, pInfo)
	}

	return &Bootstrap{
		minPeers:       c.MinPeers,
		bootstrapPeers: peers,
		host:           h,
	}, nil
}

// Start bootstrapping
func (b *Bootstrap) Start(ctx context.Context) error {
	// TODO: pre start conditions
	err := b.Bootstrap(ctx)
	return err
}

// Bootstrap thought the list of bootstrap peer's
func (b *Bootstrap) Bootstrap(ctx context.Context) error {
	var e error
	var wg sync.WaitGroup
	for _, peer := range b.bootstrapPeers {
		wg.Add(1)
		go func(peer *peerstore.PeerInfo) {
			defer wg.Done()
			bootstrapLogger.Info("start -> ", *peer)
			// TODO: Add amount judge
			if err := b.host.Connect(ctx, *peer); err != nil {
				bootstrapLogger.Info("Failed to connect to peer: ", peer)
				e = err
				return
			}
			bootstrapLogger.Info("Connected to: ", peer)
		}(peer)
	}
	wg.Wait()
	return nil
}

func BootstrapConn(p2pModule *p2p.P2P, bootstrapPeers []string) []peerstore.PeerInfo {
	ctx := context.Background()
	dht, err := dht.New(ctx, p2pModule.Host)
	if err != nil {
		bootstrapLogger.Fatal("Cant't create DHT")
	}
	// Let's connect to the bootstrap nodes first. They will tell us about the other nodes in the network.
	for _, addr := range bootstrapPeers {
		iaddr, err := ipfsaddr.ParseString(addr)
		if err != nil {
			bootstrapLogger.Info("Invalid ipfs address")
			continue
		}
		peerinfo, _ := peerstore.InfoFromP2pAddr(iaddr.Multiaddr())
		if len(peerinfo.Addrs) <= 0 {
			continue
		}
		if err := p2pModule.Host.Connect(ctx, *peerinfo); err != nil {
			bootstrapLogger.Error(err)
		} else {
			p2pModule.AddAddrToPeerstore(p2pModule.Host, addr)
			bootstrapLogger.Info("Connection established with bootstrap node: ", *peerinfo)
		}
	}

	// We use a rendezvous point "meet me here" to announce our location.
	// This is like telling your friends to meet you at the Eiffel Tower.
	rendezvousPoint, _ := cid.NewPrefixV1(cid.Raw, multihash.SHA2_256).Sum([]byte("meet me here test"))

	bootstrapLogger.Info("announcing ourselves...")
	tctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	if err := dht.Provide(tctx, rendezvousPoint, true); err != nil {
		bootstrapLogger.Error("Providers err: ", err)
	}
	bootstrapLogger.Info("searching for other peers...")
	tctx, cancel = context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	peers, err := dht.FindProviders(tctx, rendezvousPoint)
	if err != nil {
		bootstrapLogger.Error("FindProviders err: ", err)
	}
	bootstrapLogger.Info("Found", len(peers), "peers!\n")

	for _, p := range peers {
		bootstrapLogger.Info("Peer: ", p)
	}

	for _, p := range peers {
		if p.ID == p2pModule.Host.ID() || len(p.Addrs) == 0 {
			continue
		}
		s, err := p2pModule.Host.NewStream(ctx, p.ID, "/cherryCahin/1.0")
		if err != nil {
			bootstrapLogger.Error("Can't connect", err)
		} else {
			p2pModule.HandleStream(s)
			bootstrapLogger.Info("Connected to: ", p)
		}
	}

	return peers
}
