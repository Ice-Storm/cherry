package main

import (
	"context"
	"flag"
	"fmt"

	"cherrychain/common/clogging"
	"cherrychain/p2p"
	bootstrap "cherrychain/p2p/bootstrap"
	p2pUtil "cherrychain/p2p/util"
)

var bootstrapPeers = []string{
	// "/ip4/172.16.101.215/tcp/9817/ipfs/Qmb2XUn5BaMjLGE2tDyVzpK35WJ26peqXUxdHPf1FLWkGu",
	// "/ip4/104.131.131.82/tcp/4001/ipfs/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
	// "/ip4/104.236.179.241/tcp/4001/ipfs/QmSoLPppuBtQSGwKDZT2M73ULpjvfd3aZ6ha4oFGL1KrGM",
	// "/ip4/104.236.76.40/tcp/4001/ipfs/QmSoLV4Bbm51jM9C4gDYZQ9Cy3U6aXMJDAbzgu2fzaDs64",
	// "/ip4/128.199.219.111/tcp/4001/ipfs/QmSoLSafTMBsPKadTEgaXctDQVcqN88CNLHXMkTNwMKPnu",
	// "/ip4/172.16.101.215/tcp/9091/ipfs/QmR8EFE7rxsetqx7bqprPfwy5THWNH8tVxSoZjL9ah1FsE",
	// "/ip4/172.16.101.215/tcp/1121/ipfs/QmaiU2vZtq9LcSfh77LJzN4vHQKEHhRt3j343P6jCDjXrJ",
}

const protocolID = "/cherryCahin/1.0"

var mainLogger = clogging.MustGetLogger("Main")

func main() {
	port := flag.Int("sp", 3000, "listen port")
	dest := flag.String("d", "", "Dest MultiAddr String")

	flag.Parse()

	ip, _ := p2pUtil.GetLocalIP()
	ctx := context.Background()

	p2pModule := p2p.New(ctx, fmt.Sprintf("/ip4/%s/tcp/%d", ip, *port))

	if *dest != "" {
		bootstrapPeers = append(bootstrapPeers, *dest)
	}

	p2pModule.Host.SetStreamHandler(protocolID, p2pModule.HandleStream)
	fmt.Printf("./main -d /ip4/%s/tcp/%d/ipfs/%s\n", ip, *port, p2pModule.Host.ID().Pretty())

	conf := bootstrap.Config{
		BootstrapPeers: bootstrapPeers,
		MinPeers:       0,
	}

	bootstrap.Bootstrap(p2pModule, conf)

	select {}
}
