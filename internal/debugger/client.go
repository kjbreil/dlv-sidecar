package debugger

import (
	"net"
	netrpc "net/rpc"
	"net/rpc/jsonrpc"
)

const DefaultAddr = "localhost:2345"

// Client wraps a JSON-RPC connection to the Delve debugger.
type Client struct {
	conn net.Conn
	rpc  *netrpc.Client
}

// Dial connects to the Delve debugger at the given address and returns a Client.
func Dial(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &Client{
		conn: conn,
		rpc:  jsonrpc.NewClient(conn),
	}, nil
}

// Call invokes an RPC method on the Delve debugger.
func (c *Client) Call(method string, args interface{}, reply interface{}) error {
	return c.rpc.Call("RPCServer."+method, args, reply)
}

// Close closes the connection to the Delve debugger.
func (c *Client) Close() error {
	err := c.rpc.Close()
	if err2 := c.conn.Close(); err == nil {
		err = err2
	}
	return err
}
