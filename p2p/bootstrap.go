package p2p

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"cherrychain/notify"

	cid "github.com/ipfs/go-cid"
	ipfsaddr "github.com/ipfs/go-ipfs-addr"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	protocol "github.com/libp2p/go-libp2p-protocol"
	multihash "github.com/multiformats/go-multihash"
)

type Config struct {
	BootstrapPeers []string
	MinPeers       int
	NetworkID      string
	ProtocolID     protocol.ID
	Notify         *notify.Notify
}

// Bootstrap other nodes in the network.
func (n *P2P) Bootstrap(p2pModule *P2P, c Config) ([]peerstore.PeerInfo, error) {
	if c.MinPeers > len(c.BootstrapPeers) {
		return []peerstore.PeerInfo{}, errors.New(fmt.Sprintf("Too less bootstrapping nodes. Expected at least: %d, got: %d", c.MinPeers, len(c.BootstrapPeers)))
	}

	ctx := context.Background()
	dht, err := dht.New(ctx, p2pModule.Host)
	if err != nil {
		log.Fatal("Cant't create DHT")
	}

	// parallel connect to other nodes
	var wg sync.WaitGroup
	for _, addr := range c.BootstrapPeers {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			iaddr, err := ipfsaddr.ParseString(addr)
			if err != nil {
				log.Info("Invalid ipfs address")
			}
			peerinfo, _ := peerstore.InfoFromP2pAddr(iaddr.Multiaddr())
			p2pModule.Host.Peerstore().AddAddrs(peerinfo.ID, peerinfo.Addrs, peerstore.PermanentAddrTTL)
			if err := p2pModule.Host.Connect(ctx, *peerinfo); err != nil {
				log.Error(err)
			} else {
				log.Info("Connection established with bootstrap node: ", *peerinfo)
			}
		}(addr)
	}
	wg.Wait()

	peers, err := searchPeers(ctx, dht, c)
	if err != nil {
		log.Error("FindProviders err: ", err)
	}
	log.Info("Found", len(peers), "peers!\n")

	connectPeers(ctx, p2pModule, peers, c)

	return peers, nil
}

func searchPeers(ctx context.Context, dht *dht.IpfsDHT, c Config) ([]peerstore.PeerInfo, error) {
	tctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	rendezvousPoint, _ := cid.NewPrefixV1(cid.Raw, multihash.SHA2_256).Sum([]byte(c.NetworkID))
	if err := dht.Provide(tctx, rendezvousPoint, true); err != nil {
		log.Errorf("Providers err: %s. may be a genesis node or not set bootstrap node", err)
		return nil, err
	}
	log.Info("Searching for other peers...")
	tctx, cancel = context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	peers, err := dht.FindProviders(tctx, rendezvousPoint)
	if err != nil {
		log.Error("FindProviders err: ", err)
	}
	log.Info("Found", len(peers), "peers!\n")
	return peers, nil
}

func connectPeers(ctx context.Context, p2pModule *P2P, peers []peerstore.PeerInfo, c Config) {
	for _, p := range peers {
		if p.ID == p2pModule.Host.ID() || len(p.Addrs) == 0 {
			continue
		}
		s, err := p2pModule.Host.NewStream(ctx, p.ID, c.ProtocolID)
		if err != nil {
			log.Error("Can't connect", err)
		} else {
			log.Info("Connected to: ", p)
			c.Notify.SysConnected(p2pModule.Host.Network(), s)
			c.Notify.Notifee.Connected(p2pModule.Host.Network(), s.Conn())
			p2pModule.HandleStream(s)
		}
	}
}
