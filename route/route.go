package route

import (
	"context"

	host "github.com/libp2p/go-libp2p-host"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	opts "github.com/libp2p/go-libp2p-kad-dht/opts"
)

type KadDHT struct{}

func New() *KadDHT {
	return &KadDHT{}
}

func (k *KadDHT) NewDHT(ctx context.Context, h host.Host, options ...opts.Option) (*dht.IpfsDHT, error) {
	return dht.New(ctx, h, options...)
}

// func FindPeer(ctx context.Context, id peer.ID) (pstore.PeerInfo, error) {
// 	return dht.FindPeer(ctx, id)
// }
