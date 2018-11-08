package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"

	logging "cherrychain/clogging"
	"cherrychain/p2p"

	protocol "github.com/libp2p/go-libp2p-protocol"
)

var (
	bootstrapPeers = []string{
		"/ip4/0.0.0.0/tcp/9098/ipfs/QmZm5dYY6DSF5uXduZk99PtbmxRsNQ9vJad1RaR33rxhij",
		// "/ip4/104.131.131.82/tcp/4001/ipfs/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
		// "/ip4/104.236.179.241/tcp/4001/ipfs/QmSoLPppuBtQSGwKDZT2M73ULpjvfd3aZ6ha4oFGL1KrGM",
		// "/ip4/104.236.76.40/tcp/4001/ipfs/QmSoLV4Bbm51jM9C4gDYZQ9Cy3U6aXMJDAbzgu2fzaDs64",
		// "/ip4/128.199.219.111/tcp/4001/ipfs/QmSoLSafTMBsPKadTEgaXctDQVcqN88CNLHXMkTNwMKPnu",
		// "/ip4/178.62.158.247/tcp/4001/ipfs/QmSoLer265NRgSp2LA3dPaeykiS1J6DifTC88f5uVQKNAd",
		//"/ip4/0.0.0.0/tcp/4001/p2p/QmdSyhb8eR9dDSR5jjnRoTDBwpBCSAjT7WueKJ9cQArYoA",
	}
	log    = logging.MustGetLogger("MAIN")
	isRoot = true
)

const (
	ip         = "0.0.0.0"
	protocolID = "/cherryCahin/1.0"
	networkID  = "cherry-test"
)

func main() {
	port := flag.Int("sp", 3000, "listen port")
	dest := flag.String("d", "", "Dest MultiAddr String")
	flag.Parse()

	if *dest != "" {
		isRoot = false
		bootstrapPeers = append(bootstrapPeers, *dest)
	}

	ctx, cancel := context.WithCancel(context.Background())
	p2pNetwork := p2p.New(ctx, fmt.Sprintf("/ip4/%s/tcp/%d", ip, *port), isRoot)

	if err := p2pNetwork.StartSysEventLoop(ctx); err != nil {
		cancel()
	}

	pID := protocol.ID(protocolID)
	p2pNetwork.Host.SetStreamHandler(pID, p2pNetwork.HandleStream)

	log.Notice(fmt.Sprintf("./main -d /ip4/%s/tcp/%d/ipfs/%s \n", ip, *port, p2pNetwork.Host.ID().Pretty()))

	conf := p2p.Config{
		BootstrapPeers: bootstrapPeers,
		MinPeers:       0,
		NetworkID:      networkID,
		ProtocolID:     pID,
		Notify:         p2pNetwork.Notify,
	}

	if !isRoot {
		if _, err := p2pNetwork.Bootstrap(p2pNetwork, conf); err == nil {
			go writeData(p2pNetwork)
			go readData(p2pNetwork)
		}
	}

	select {}
}

func writeData(network *p2p.P2P) {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		network.Write([]byte(sendData))
	}
}

func readData(network *p2p.P2P) {
	for {
		cap := make([]byte, 1000)
		network.Read(cap)
		fmt.Printf("\x1b[32m%s\x1b[0m> ", string(cap))
	}
}
