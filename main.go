package main

import (
	"context"
	"flag"
	"fmt"

	"cherrychain/common/clogging"
	"cherrychain/p2p"
	p2pUtil "cherrychain/p2p/util"

	host "github.com/libp2p/go-libp2p-host"
)

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
		peerID := p2pModule.AddAddrToPeerstore(p2pModule.Host, *dest)
		s, err := p2pModule.Host.NewStream(context.Background(), peerID, protocolID)
		if err != nil {
			mainLogger.Fatal("Can't connect %s", *dest)
		}
		p2pModule.HandleStream(s)
		select {}
	}
}
