package localrelay

import (
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

func TestConnPoolBasic(t *testing.T) {
	conns := []net.Conn{}
	connAmount := 50
	relay := New("test-relay", "127.0.0.1:23838", "127.0.0.1:23838", io.Discard)

	for i := 0; i < connAmount; i++ {
		conn := &net.TCPConn{}

		conns = append(conns, conn)
		relay.storeConn(conn)
	}

	for i := 0; i < connAmount; i++ {
		relay.popConn(conns[i])
	}

	if len(relay.connPool) != 0 {
		t.Fatal("connPool is not empty")
	}
}

func TestConnPool(t *testing.T) {
	// create channel to receive errors from another goroutine
	errCh := make(chan error)
	go startTCPServer(errCh)

	// wait for error or nil error indicating server launched fine
	if err := <-errCh; err != nil {
		t.Fatal(err)
	}

	relay := New("test-relay", "127.0.0.1:23838", "127.0.0.1:23838", io.Discard)

	wg := sync.WaitGroup{}

	// open 10 conns and append to the conn pool
	for i := 0; i < 10; i++ {
		wg.Add(1)

		conn, err := net.Dial("tcp", "127.0.0.1:23838")
		if err != nil {
			t.Fatal(err)
		}

		relay.storeConn(conn)

		// handle conn
		go func(conn net.Conn, i int) {
			for {
				time.Sleep(time.Millisecond * (10 * time.Duration(i)))
				_, err := conn.Write([]byte("test"))
				if err != nil {
					relay.popConn(conn)

					for _, c := range relay.connPool {
						if c.Conn == conn {
							t.Fatal("correct conn was not removed")
						}
					}

					wg.Done()
					return
				}
			}
		}(conn, i)
	}

	wg.Wait()
}

func startTCPServer(errCh chan error) {
	l, err := net.Listen("tcp", ":23838")
	if err != nil {
		errCh <- err
		return
	}

	errCh <- nil

	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}

		// handle conn with echo server
		go func(conn net.Conn) {
			for i := 0; i <= 5; i++ {
				buf := make([]byte, 1048)
				n, err := conn.Read(buf)
				if err != nil {
					conn.Close()
					return
				}

				conn.Write(buf[:n])
			}

			// close conn after 5 messages
			conn.Close()
		}(conn)
	}
}
