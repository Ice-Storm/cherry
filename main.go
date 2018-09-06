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
	"/ip4/172.16.101.217/tcp/9816/ipfs/QmTLP33p9FgZWhcPDYZqYfehvxmefn4pTGC517PaCGM8YX",
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

		peers := bootstrap.BootstrapConn(p2pModule.Host, bootstrapPeers)
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
