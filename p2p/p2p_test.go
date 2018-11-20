package p2p

import (
	"context"
	"fmt"
	"testing"
)

func TestNewHost(t *testing.T) {
	p2pRoot, err := makeRandomHost(t, 9000, true)
	if err != nil {
		t.Fatal(err)
	}
	p2pRoot.Host.Close()

	p2p, err := makeRandomHost(t, 9000, false)
	if err != nil {
		t.Fatal(err)
	}
	p2p.Host.Close()
}

func makeRandomHost(t *testing.T, port int, isRoot bool) (*P2P, error) {
	ctx := context.Background()
	return New(ctx, fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port), isRoot)
}
