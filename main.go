package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"cherrychain/commands"
	"cherrychain/common/clogging"
	config "cherrychain/config"
	"cherrychain/p2p"
	"cherrychain/p2p/bootstrap"
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

var mainLogger = clogging.MustGetLogger("MAIN")

func main() {
	cflag := commands.CommandInit()
	fconf, _ := config.Load(cflag.Fconf)
	ip, _ := p2pUtil.GetLocalIP()
	ctx := context.Background()

	p2pModule := p2p.New(ctx, fmt.Sprintf("/ip4/%s/tcp/%d", ip, cflag.Port))
	p2pModule.StartSysEventLoop()

	if cflag.Dest != "" {
		bootstrapPeers = append(bootstrapPeers, cflag.Dest)
	}

	p2pModule.Host.SetStreamHandler(fconf.ProtocolID, p2pModule.HandleStream)
	mainLogger.Notice(fmt.Sprintf("./main -d /ip4/%s/tcp/%d/ipfs/%s -f cherry\n", ip, cflag.Port, p2pModule.Host.ID().Pretty()))

	conf := bootstrap.Config{
		BootstrapPeers: bootstrapPeers,
		MinPeers:       0,
		NetworkID:      fconf.NetworkID,
		ProtocolID:     fconf.ProtocolID,
		Notify:         p2pModule.Notify,
	}

	bootstrap.Bootstrap(p2pModule, conf)

	stdReader := bufio.NewReader(os.Stdin)
	go func() {
		for {
			fmt.Print("> ")
			sendData, err := stdReader.ReadString('\n')
			if err != nil {
				panic(err)
			}
			p2pModule.Write([]byte(sendData))
		}
	}()

	go func() {
		for {
			p2pModule.Read()
		}
	}()

	select {}
}
