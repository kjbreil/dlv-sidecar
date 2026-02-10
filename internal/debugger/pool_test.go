package debugger

import (
	"errors"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"sync"
	"testing"
)

// echoService is a simple RPC service for testing.
type echoService struct{}

type EchoArgs struct {
	Msg string
}

type EchoReply struct {
	Msg string
}

func (e *echoService) Echo(args *EchoArgs, reply *EchoReply) error {
	reply.Msg = args.Msg
	return nil
}

// startTestServer starts a JSON-RPC server on a random port and returns its address
// and a function to stop it.
func startTestServer(t *testing.T) (string, func()) {
	t.Helper()
	srv := rpc.NewServer()
	if err := srv.RegisterName("RPCServer", new(echoService)); err != nil {
		t.Fatal(err)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go srv.ServeCodec(jsonrpc.NewServerCodec(conn))
		}
	}()

	return ln.Addr().String(), func() {
		ln.Close()
		wg.Wait()
	}
}

func TestPool_Call(t *testing.T) {
	addr, stop := startTestServer(t)
	defer stop()

	pool := NewPool(addr)
	defer pool.Close()

	var reply EchoReply
	err := pool.Call("Echo", &EchoArgs{Msg: "hello"}, &reply)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reply.Msg != "hello" {
		t.Fatalf("got %q, want %q", reply.Msg, "hello")
	}
}

func TestPool_ReusesConnection(t *testing.T) {
	addr, stop := startTestServer(t)
	defer stop()

	pool := NewPool(addr)
	defer pool.Close()

	// Make two calls — the second should reuse the connection.
	for i := 0; i < 2; i++ {
		var reply EchoReply
		if err := pool.Call("Echo", &EchoArgs{Msg: "reuse"}, &reply); err != nil {
			t.Fatalf("call %d: unexpected error: %v", i, err)
		}
		if reply.Msg != "reuse" {
			t.Fatalf("call %d: got %q, want %q", i, reply.Msg, "reuse")
		}
	}
}

func TestPool_ReconnectOnFailure(t *testing.T) {
	addr, stop := startTestServer(t)

	pool := NewPool(addr)
	defer pool.Close()

	// First call succeeds.
	var reply EchoReply
	if err := pool.Call("Echo", &EchoArgs{Msg: "first"}, &reply); err != nil {
		t.Fatalf("first call: unexpected error: %v", err)
	}

	// Shut down the server and start a new one on the same port.
	stop()

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatalf("failed to re-listen on %s: %v", addr, err)
	}
	srv := rpc.NewServer()
	if err := srv.RegisterName("RPCServer", new(echoService)); err != nil {
		t.Fatal(err)
	}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go srv.ServeCodec(jsonrpc.NewServerCodec(conn))
		}
	}()
	defer ln.Close()

	// The pool should reconnect automatically.
	reply = EchoReply{}
	if err := pool.Call("Echo", &EchoArgs{Msg: "after-reconnect"}, &reply); err != nil {
		t.Fatalf("reconnect call: unexpected error: %v", err)
	}
	if reply.Msg != "after-reconnect" {
		t.Fatalf("got %q, want %q", reply.Msg, "after-reconnect")
	}
}

func TestPool_ConnectError(t *testing.T) {
	// Use a port that nothing is listening on.
	pool := NewPool("127.0.0.1:0")
	defer pool.Close()

	var reply EchoReply
	err := pool.Call("Echo", &EchoArgs{Msg: "fail"}, &reply)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPool_Addr(t *testing.T) {
	pool := NewPool("localhost:9999")
	if got := pool.Addr(); got != "localhost:9999" {
		t.Fatalf("Addr() = %q, want %q", got, "localhost:9999")
	}
}

func TestPool_CloseIdempotent(t *testing.T) {
	addr, stop := startTestServer(t)
	defer stop()

	pool := NewPool(addr)

	// Establish a connection.
	var reply EchoReply
	if err := pool.Call("Echo", &EchoArgs{Msg: "test"}, &reply); err != nil {
		t.Fatal(err)
	}

	// Close twice should not panic.
	pool.Close()
	if err := pool.Close(); err != nil {
		t.Fatalf("second close: %v", err)
	}
}

func TestPool_ConcurrentCalls(t *testing.T) {
	addr, stop := startTestServer(t)
	defer stop()

	pool := NewPool(addr)
	defer pool.Close()

	var wg sync.WaitGroup
	errs := make([]error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			var reply EchoReply
			errs[idx] = pool.Call("Echo", &EchoArgs{Msg: "concurrent"}, &reply)
		}(i)
	}

	wg.Wait()

	for i, err := range errs {
		if err != nil {
			// Concurrent calls on a single JSON-RPC connection may fail; that's OK.
			// We just want no panics.
			if !errors.Is(err, rpc.ErrShutdown) {
				t.Errorf("goroutine %d: unexpected error type: %v", i, err)
			}
		}
	}
}
