package p2p

import (
	"bufio"
	"context"

	"cherrychain/common/clogging"

	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-host"
	multiaddr "github.com/multiformats/go-multiaddr"
)

var p2pLogger = clogging.MustGetLogger("P2P")

type P2P struct{}

func New() *P2P {
	return &P2P{}
}

func (n *P2P) GenesisNode(genesisMultiAddr string) (host.Host, error) {
	sourceMultiAddr, _ := multiaddr.NewMultiaddr(genesisMultiAddr)
	p2pLogger.Info("New libp2p object")
	host, _ := libp2p.New(
		context.Background(),
		libp2p.ListenAddrs(sourceMultiAddr),
	)
	return host, nil
}

func (n *P2P) ReadString(IO interface{}) (string, error) {
	switch stream := IO.(type) {
	case *bufio.ReadWriter:
		return stream.ReadString('\n')
	case *bufio.Reader:
		return stream.ReadString('\n')
	default:
		return "", nil
	}
}

func (n *P2P) WriteBytes(stream *bufio.Writer, str []byte) error {
	n.writeString(stream, string(append(str, '\n')))
	return nil
}

func (n *P2P) writeString(stream *bufio.Writer, str string) error {
	stream.WriteString(str)
	err := stream.Flush()
	return err
}
