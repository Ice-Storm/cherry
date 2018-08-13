package p2p

import (
	"errors"

	peer "github.com/libp2p/go-libp2p-peer"
)

type PeerStore struct {
	Store []PeerDiscovery
}

type PeerDiscovery struct {
	ID       peer.ID
	Protocol string
	Port     uint16
}

func NewPeerStore() *PeerStore {
	return &PeerStore{make([]PeerDiscovery, 20)}
}

func (p *PeerStore) push(peerInfo PeerDiscovery) ([]PeerDiscovery, error) {
	for _, val := range p.Store {
		if val.ID == peerInfo.ID {
			return nil, errors.New("Peer store has repeat peer id.")
		}
	}
	return append(p.Store, peerInfo), nil
}
