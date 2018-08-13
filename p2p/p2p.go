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

func (n *P2P) WriteString(stream *bufio.ReadWriter, str string) error {
	stream.WriteString(str)
	err := stream.Flush()
	return err
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

func (n *P2P) SwapPeerInfo(stream *bufio.ReadWriter, str []byte) error {
	n.WriteString(stream, string(append(str, '\n')))
	return nil
}

func (n *P2P) WriteBytes(IO interface{}, data []byte) (int, error) {
	switch stream := IO.(type) {
	case *bufio.Writer:
		n, err := stream.Write(append(data, '\n'))
		stream.Flush()
		return n, err
	default:
		return -1, nil
	}
}

func (n *P2P) ReadBytes(IO interface{}) ([]byte, error) {
	switch stream := IO.(type) {
	case *bufio.Reader:
		return stream.ReadBytes('\n')
	default:
		return nil, nil
	}
}
