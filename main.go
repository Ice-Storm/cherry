package main

import (
	"context"
	"flag"
	"fmt"

	"cherrychain/common/clogging"
	"cherrychain/p2p"
	bootstrap "cherrychain/p2p/bootstrap"
	p2pUtil "cherrychain/p2p/util"

	host "github.com/libp2p/go-libp2p-host"
)

var bootstrapPeers = []string{
	"/ip4/128.199.219.111/tcp/4001/ipfs/QmSoLSafTMBsPKadTEgaXctDQVcqN88CNLHXMkTNwMKPnu",
	"/ip4/172.16.101.215/tcp/9816/ipfs/QmNk7mn6viz4quXVyiVPoXU1MHhXs6tkkoS9uSDNNVSSvy",
	"/ip4/172.16.101.215/tcp/9817/ipfs/QmUPnXrvjNQan2nLRHMv8nj4v4Maj8s47Yzj7hvE83vGYJ",
	"/ip4/178.62.158.247/tcp/4001/ipfs/QmSoLer265NRgSp2LA3dPaeykiS1J6DifTC88f5uVQKNAd",
}

const protocolID = "/cherryCahin/1.0"

var mainLogger = clogging.MustGetLogger("Main")

type network struct {
	p2p     *p2p.P2P
	peerDis *p2p.PeerDiscovery
	Host    host.Host
}

func main() {
	port := flag.Int("sp", 3000, "listen port")
	dest := flag.String("d", "", "Dest MultiAddr String")

	flag.Parse()

	ip, _ := p2pUtil.GetLocalIP()

	ctx := context.Background()
	p2pModule := p2p.New(ctx, fmt.Sprintf("/ip4/%s/tcp/%d", ip, *port))

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
		fmt.Printf("./main -d /ip4/%s/tcp/%d/ipfs/%s\n", ip, *port, p2pModule.Host.ID().Pretty())
		p2pModule.AddAddrToPeerstore(p2pModule.Host, *dest)
		bootstrap.BootstrapConn(p2pModule, bootstrapPeers)
		select {}
	}
}
