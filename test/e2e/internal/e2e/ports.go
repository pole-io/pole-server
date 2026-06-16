//go:build e2e
// +build e2e

package e2e

import (
	"fmt"
	"net"
	"testing"
)

type PortAllocator struct {
	next int
}

func NewPortAllocator(base int) *PortAllocator {
	return &PortAllocator{next: base}
}

func (p *PortAllocator) Next(t *testing.T) int {
	t.Helper()
	for port := p.next; port < p.next+200; port++ {
		addr := fmt.Sprintf("127.0.0.1:%d", port)
		ln, err := net.Listen("tcp", addr)
		if err == nil {
			_ = ln.Close()
			p.next = port + 1
			return port
		}
	}
	t.Fatalf("no available tcp port from %d", p.next)
	return 0
}
