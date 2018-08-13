package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"sync"

	"cherrychain/p2p"

	host "github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	multiaddr "github.com/multiformats/go-multiaddr"
)

const protocolID = "/cherryCahin/1.0"

var mutex = &sync.Mutex{}

type network struct {
	p2p *p2p.P2P
}

func (p *network) handleStream(s net.Stream) {
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	go p.readData(rw)
	go p.writeData(rw)
}

func addAddrToPeerstore(h host.Host, addr string) peer.ID {
	cherryAddr, err := multiaddr.NewMultiaddr(addr)

	if err != nil {
		log.Fatalln(err)
	}

	pid, err := cherryAddr.ValueForProtocol(multiaddr.P_IPFS)

	if err != nil {
		log.Fatalln(err)
	}

	peerid, err := peer.IDB58Decode(pid)

	if err != nil {
		log.Fatalln(err)
	}

	targetPeerAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ipfs/%s", peer.IDB58Encode(peerid)))
	targetAddr := cherryAddr.Decapsulate(targetPeerAddr)

	fmt.Printf(pid)

	h.Peerstore().AddAddr(peerid, targetAddr, peerstore.PermanentAddrTTL)
	return peerid
}

func (p *network) readData(rw *bufio.ReadWriter) {
	for {
		str, _ := p.p2p.ReadString(rw)
		if str == "" {
			return
		}
		fmt.Printf(str)
	}
}

func (p *network) writeData(rw *bufio.ReadWriter) {
	stdReader := bufio.NewReader(os.Stdin)
	go func() {
		for {
			time.Sleep(5 * time.Second)
			mutex.Lock()
			p.p2p.CheckAlive(rw, "sendData\n")
			mutex.Unlock()
		}
	}()

	for {
		sendData, err := p.p2p.ReadString(stdReader)
		if err != nil {
			panic(err)
		}
		mutex.Lock()
		p.p2p.WriteString(rw, sendData)
		mutex.Unlock()
	}
}

func main() {
	port := flag.Int("sp", 3000, "listen port")
	dest := flag.String("d", "", "Dest MultiAddr String")

	flag.Parse()

	var p2pNetwork = &network{p2p.New()}
	host, err := p2pNetwork.p2p.GenesisNode(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", *port))

	if err != nil {
		panic(err)
	}

	if *dest == "" {
		host.SetStreamHandler(protocolID, p2pNetwork.handleStream)
		fmt.Printf("./main -d /ip4/127.0.0.1/tcp/%d/ipfs/%s", *port, host.ID().Pretty())
		<-make(chan struct{})
	} else {
		host.SetStreamHandler(protocolID, p2pNetwork.handleStream)
		fmt.Printf("./main -d /ip4/127.0.0.1/tcp/%d/ipfs/%s", *port, host.ID().Pretty())
		fmt.Printf("\n")
		peerID := addAddrToPeerstore(host, *dest)
		s, err := host.NewStream(context.Background(), peerID, protocolID)
		if err != nil {
			panic(err)
		}
		p2pNetwork.handleStream(s)
		select {}
	}

}
