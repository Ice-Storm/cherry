package p2p

import (
	"errors"
)

type PeerStore struct {
	Store []PeerDiscovery
}

type PeerDiscovery struct {
	ID       string
	Protocol string
	Port     uint16
	IP       string
}

func NewPeerStore() *PeerStore {
	return &PeerStore{make([]PeerDiscovery, 0)}
}

func (p *PeerStore) Push(peerInfo PeerDiscovery) ([]PeerDiscovery, error) {
	for _, val := range p.Store {
		if val.ID == peerInfo.ID {
			return nil, errors.New("Peer store has repeat peer id.")
		}
	}
	p.Store = append(p.Store, peerInfo)
	return p.Store, nil
}
