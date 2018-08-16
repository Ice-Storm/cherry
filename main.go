package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"sync"

	"cherrychain/common/clogging"
	"cherrychain/p2p"

	host "github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	multiaddr "github.com/multiformats/go-multiaddr"
)

const protocolID = "/cherryCahin/1.0"

var mutex = &sync.Mutex{}

var mainLogger = clogging.MustGetLogger("Main")

var pStore = p2p.NewPeerStore()

type network struct {
	p2p     *p2p.P2P
	peerDis *p2p.PeerDiscovery
	Host    host.Host
}

func (p *network) handleStream(s net.Stream) {
	p.SwapPeersInfo(s)
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

	h.Peerstore().AddAddr(peerid, targetAddr, peerstore.PermanentAddrTTL)
	return peerid
}

func (p *network) SwapPeersInfo(s net.Stream) {
	p.readData(s)
	p.writeData(s)
}

func (p *network) readData(s net.Stream) {
	rw := bufio.NewReader(s)
	go func() {
		for {
			str, _ := p.p2p.ReadString(rw)
			if str == "" {
				return
			}

			var peerInfo p2p.PeerDiscovery
			json.Unmarshal([]byte(str), &peerInfo)
			_, err := pStore.Push(peerInfo)

			if err != nil {
				mainLogger.Info("Failed push new peer info")
			} else {
				fmt.Printf("\nPeerInfo %v \n", peerInfo)
				fmt.Printf("Host %s \n", p.Host.ID().Pretty())

				if peerInfo.ID != p.Host.ID().Pretty() {
					fmt.Printf("********")
					fmt.Printf("\n/ip4/0.0.0.0/tcp/%d/ipfs/%s\n", peerInfo.Port, peerInfo.ID)
					p.attach(peerInfo.Port, peerInfo.ID)
				}
			}
			fmt.Printf("\nlen : %d\n", len(pStore.Store))
		}
	}()
}

func (p *network) writeData(s net.Stream) {
	rw := bufio.NewWriter(s)
	go func() {
		for {
			for i := 0; i < len(pStore.Store); i++ {
				time.Sleep(5 * time.Second)
				mutex.Lock()
				data, err := json.Marshal(pStore.Store[i])
				if err != nil {
					mainLogger.Fatal("Failed marshal peer store")
				}
				p.p2p.WriteBytes(rw, data)
				mutex.Unlock()
				if i > len(pStore.Store) {
					i = 0
				}
			}
		}
	}()
}

func (p *network) attach(port uint16, hostID string) error {
	peerID := addAddrToPeerstore(p.Host, fmt.Sprintf("/ip4/0.0.0.0/tcp/%d/ipfs/%s", port, hostID))
	s, err := p.Host.NewStream(context.Background(), peerID, protocolID)
	if err != nil {
		panic(err)
	}
	p.handleStream(s)
	return err
}

func main() {
	port := flag.Int("sp", 3000, "listen port")
	dest := flag.String("d", "", "Dest MultiAddr String")

	flag.Parse()

	var p2pNetwork = &network{
		p2p: p2p.New(),
	}
	host, err := p2pNetwork.p2p.GenesisNode(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", *port))

	p2pNetwork.Host = host
	p2pNetwork.peerDis = &p2p.PeerDiscovery{
		ID:       host.ID().Pretty(),
		Protocol: protocolID,
		Port:     uint16(*port),
	}

	if err != nil {
		panic(err)
	}

	pStore.Push(*p2pNetwork.peerDis)

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
			mainLogger.Fatal("Can't connect", fmt.Sprintf("/ip4/0.0.0.0/tcp/%d/ipfs/%s", *port, host.ID().Pretty()))
		}
		p2pNetwork.handleStream(s)
		select {}
	}

}
