package bootstrap

import (
	"cherrychain/common/clogging"
	"context"
	"math/rand"
	"strconv"
	"time"

	cid "github.com/ipfs/go-cid"
	ipfsaddr "github.com/ipfs/go-ipfs-addr"
	host "github.com/libp2p/go-libp2p-host"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	multihash "github.com/multiformats/go-multihash"
)

var bootstrapLogger = clogging.MustGetLogger("P2P")

func BootstrapConn(host host.Host, bootstrapPeers []string) []peerstore.PeerInfo {
	ctx := context.Background()
	dht, err := dht.New(ctx, host)
	if err != nil {
		panic(err)
	}
	// Let's connect to the bootstrap nodes first. They will tell us about the other nodes in the network.
	for _, addr := range bootstrapPeers {
		iaddr, _ := ipfsaddr.ParseString(addr)
		peerinfo, _ := peerstore.InfoFromP2pAddr(iaddr.Multiaddr())
		if err := host.Connect(ctx, *peerinfo); err != nil {
			bootstrapLogger.Error(err)
		} else {
			bootstrapLogger.Info("Connection established with bootstrap node: ", *peerinfo)
		}
	}
	// We use a rendezvous point "meet me here" to announce our location.
	// This is like telling your friends to meet you at the Eiffel Tower.
	rand.Seed(time.Now().Unix())
	rendezvousPoint, _ := cid.NewPrefixV1(cid.Raw, multihash.SHA2_256).Sum([]byte(strconv.Itoa(rand.Intn(100))))
	bootstrapLogger.Info("announcing ourselves...")
	tctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	if err := dht.Provide(tctx, rendezvousPoint, true); err != nil {
		bootstrapLogger.Info("Providers err")
		panic(err)
	}
	bootstrapLogger.Info("searching for other peers...")
	tctx, cancel = context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	peers, err := dht.FindProviders(tctx, rendezvousPoint)
	if err != nil {
		bootstrapLogger.Info("FindProviders err")
		panic(err)
	}
	bootstrapLogger.Info("Found", len(peers), "peers!\n")
	for _, p := range peers {
		bootstrapLogger.Info("Peer: ", p)
	}
	return peers
}
