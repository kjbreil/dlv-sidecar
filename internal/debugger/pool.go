package debugger

import (
	"fmt"
	"sync"
)

// Pool maintains a persistent JSON-RPC client connection to a Delve debugger,
// reconnecting automatically on failure.
type Pool struct {
	addr   string
	mu     sync.Mutex
	client *Client
}

// NewPool creates a connection pool for the given Delve debugger address.
func NewPool(addr string) *Pool {
	return &Pool{addr: addr}
}

// Addr returns the address this pool connects to.
func (p *Pool) Addr() string {
	return p.addr
}

// getClient returns an existing healthy connection or dials a new one.
func (p *Pool) getClient() (*Client, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.client != nil {
		return p.client, nil
	}

	c, err := Dial(p.addr)
	if err != nil {
		return nil, err
	}
	p.client = c
	return p.client, nil
}

// discard closes and removes the current connection so the next call redials.
func (p *Pool) discard() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.client != nil {
		p.client.Close()
		p.client = nil
	}
}

// Call invokes an RPC method using the pooled connection.
// On failure it retries once with a fresh connection.
func (p *Pool) Call(method string, args interface{}, reply interface{}) error {
	c, err := p.getClient()
	if err != nil {
		return fmt.Errorf("connect to %s: %w", p.addr, err)
	}

	err = c.Call(method, args, reply)
	if err == nil {
		return nil
	}

	// Connection may be stale; discard and retry once.
	p.discard()

	c, err = p.getClient()
	if err != nil {
		return fmt.Errorf("reconnect to %s: %w", p.addr, err)
	}

	return c.Call(method, args, reply)
}

// Close closes the pooled connection.
func (p *Pool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.client == nil {
		return nil
	}
	err := p.client.Close()
	p.client = nil
	return err
}
