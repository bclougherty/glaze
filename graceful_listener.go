package glaze

import (
	"net"
	"os"
)

type signal struct{}

type gracefulListener struct {
	net.Listener
	conn gracefulConn

	stop    chan signal
	stopped bool
}

// File is a convenience method for getting a gracefulListener's file descriptor
func (gl gracefulListener) File() *os.File {
	tl := gl.Listener.(*net.TCPListener)
	fl, _ := tl.File()
	return fl
}

// Accept accepts a new connection as net.Listener.Accept, but also tracks connection count so that we know how many open connections there are
func (gl gracefulListener) Accept() (net.Conn, error) {
	conn, err := gl.Listener.Accept()
	if err != nil {
		return nil, err
	}

	// Wrap the returned connection, so that we can observe when
	// it is closed.
	gl.conn = newGracefulConn(conn)

	return conn, nil
}
