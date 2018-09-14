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
	dht "github.com/libp2p/go-libp2p-kad-dht"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	multihash "github.com/multiformats/go-multihash"
)

var bootstrapLogger = clogging.MustGetLogger("bootstrap")
var networkID = "meet me here test"

type Config struct {
	BootstrapPeers []string
	MinPeers       int
	NetworkID      string
}

func Bootstrap(p2pModule *p2p.P2P, c Config) ([]peerstore.PeerInfo, error) {
	if c.MinPeers > len(c.BootstrapPeers) {
		return []peerstore.PeerInfo{}, errors.New(fmt.Sprintf("Too less bootstrapping nodes. Expected at least: %d, got: %d", c.MinPeers, len(c.BootstrapPeers)))
	}

	ctx := context.Background()
	dht, err := dht.New(ctx, p2pModule.Host)
	if err != nil {
		bootstrapLogger.Fatal("Cant't create DHT")
	}

	// Let's connect to the bootstrap nodes first. They will tell us about the other nodes in the network.
	var wg sync.WaitGroup
	for _, addr := range c.BootstrapPeers {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			iaddr, err := ipfsaddr.ParseString(addr)
			if err != nil {
				bootstrapLogger.Info("Invalid ipfs address")
			}
			peerinfo, _ := peerstore.InfoFromP2pAddr(iaddr.Multiaddr())
			if err := p2pModule.Host.Connect(ctx, *peerinfo); err != nil {
				bootstrapLogger.Error(err)
			} else {
				bootstrapLogger.Info("Connection established with bootstrap node: ", *peerinfo)
			}
		}(addr)
	}
	wg.Wait()

	// We use a rendezvous point "meet me here" to announce our location.
	// This is like telling your friends to meet you at the Eiffel Tower.
	rendezvousPoint, _ := cid.NewPrefixV1(cid.Raw, multihash.SHA2_256).Sum([]byte(c.NetworkID))

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

	return peers, nil
}
