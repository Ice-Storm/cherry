package bootstrap

import (
	"cherrychain/common/clogging"
	"cherrychain/p2p"
	"context"
	"time"

	cid "github.com/ipfs/go-cid"
	ipfsaddr "github.com/ipfs/go-ipfs-addr"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	multihash "github.com/multiformats/go-multihash"
)

var bootstrapLogger = clogging.MustGetLogger("P2P")
var preFix = "meet me here test"

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
	rendezvousPoint, _ := cid.NewPrefixV1(cid.Raw, multihash.SHA2_256).Sum([]byte(preFix))

	bootstrapLogger.Info("announcing ourselves...")
	tctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	if err := dht.Provide(tctx, rendezvousPoint, true); err != nil {
		bootstrapLogger.Fatal("Providers err")
	}
	bootstrapLogger.Info("searching for other peers...")
	tctx, cancel = context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	peers, err := dht.FindProviders(tctx, rendezvousPoint)
	if err != nil {
		bootstrapLogger.Fatal("FindProviders err")
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
