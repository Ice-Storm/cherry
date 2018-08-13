package p2p

import (
	"errors"

	peer "github.com/libp2p/go-libp2p-peer"
)

type PeerStore struct {
	Store []peer.ID
}

func NewPeerStore() *PeerStore {
	return &PeerStore{make([]peer.ID, 20)}
}

func (p *PeerStore) push(id peer.ID) ([]peer.ID, error) {
	for _, val := range p.Store {
		if val == id {
			return nil, errors.New("Peer store has repeat peer id.")
		}
	}
	return append(p.Store, id), nil
}
