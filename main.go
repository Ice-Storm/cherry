package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"cherrychain/common/clogging"
	"cherrychain/p2p"
	p2pUtil "cherrychain/p2p/util"

	cid "github.com/ipfs/go-cid"
	ipfsaddr "github.com/ipfs/go-ipfs-addr"
	host "github.com/libp2p/go-libp2p-host"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	multihash "github.com/multiformats/go-multihash"
)

var bootstrapPeers = []string{
	"/ip4/104.236.76.40/tcp/4001/ipfs/QmSoLV4Bbm51jM9C4gDYZQ9Cy3U6aXMJDAbzgu2fzaDs64",
	"/ip4/128.199.219.111/tcp/4001/ipfs/QmSoLSafTMBsPKadTEgaXctDQVcqN88CNLHXMkTNwMKPnu",
	"/ip4/178.62.158.247/tcp/4001/ipfs/QmSoLer265NRgSp2LA3dPaeykiS1J6DifTC88f5uVQKNAd",
}

const protocolID = "/cherryCahin/1.0"

var mainLogger = clogging.MustGetLogger("Main")

type network struct {
	p2p     *p2p.P2P
	peerDis *p2p.PeerDiscovery
	Host    host.Host
}

func bootstrapConn(host host.Host) []peerstore.PeerInfo {
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
			mainLogger.Error(err)
		} else {
			mainLogger.Info("Connection established with bootstrap node: ", *peerinfo)
		}
	}
	// We use a rendezvous point "meet me here" to announce our location.
	// This is like telling your friends to meet you at the Eiffel Tower.
	rand.Seed(time.Now().Unix())
	rendezvousPoint, _ := cid.NewPrefixV1(cid.Raw, multihash.SHA2_256).Sum([]byte(strconv.Itoa(rand.Intn(100))))
	fmt.Println("announcing ourselves...")
	tctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	if err := dht.Provide(tctx, rendezvousPoint, true); err != nil {
		mainLogger.Info("Providers err")
		panic(err)
	}
	fmt.Println("searching for other peers...")
	tctx, cancel = context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	peers, err := dht.FindProviders(tctx, rendezvousPoint)
	if err != nil {
		mainLogger.Info("FindProviders err")
		panic(err)
	}
	fmt.Printf("Found %d peers!\n", len(peers))
	for _, p := range peers {
		fmt.Println("Peer: ", p)
	}
	return peers
}

func main() {
	port := flag.Int("sp", 3000, "listen port")
	dest := flag.String("d", "", "Dest MultiAddr String")

	flag.Parse()

	ip, _ := p2pUtil.GetLocalIP()

	p2pModule := p2p.New(fmt.Sprintf("/ip4/%s/tcp/%d", ip, *port))

	var p2pNetwork = &network{
		p2p: p2pModule,
		peerDis: &p2p.PeerDiscovery{
			ID:       p2pModule.Host.ID().Pretty(),
			Protocol: protocolID,
			Port:     uint16(*port),
			IP:       ip,
		},
	}

	p2pModule.PeerStore.Push(*p2pNetwork.peerDis)

	if *dest == "" {
		p2pModule.Host.SetStreamHandler(protocolID, p2pModule.HandleStream)
		fmt.Printf("./main -d /ip4/%s/tcp/%d/ipfs/%s", ip, *port, p2pModule.Host.ID().Pretty())
		<-make(chan struct{})
	} else {
		p2pModule.Host.SetStreamHandler(protocolID, p2pModule.HandleStream)
		fmt.Printf("./main -d /ip4/%s/tcp/%d/ipfs/%s", ip, *port, p2pModule.Host.ID().Pretty())
		fmt.Printf("\n")
		p2pModule.AddAddrToPeerstore(p2pModule.Host, *dest)

		peers := bootstrapConn(p2pModule.Host)
		for _, p := range peers {
			if p.ID == p2pModule.Host.ID() || len(p.Addrs) == 0 {
				continue
			}

			s, err := p2pModule.Host.NewStream(context.Background(), p.ID, protocolID)
			if err != nil {
				mainLogger.Error("Can't connect %s", err)
			}

			p2pModule.HandleStream(s)
			fmt.Println("Connected to: ", p)

		}
		select {}
	}
}
