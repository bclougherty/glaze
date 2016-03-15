package glaze

import "net"

type gracefulConn struct {
	net.Conn

	numCons counter
}

func newGracefulConn(c net.Conn) gracefulConn {
	gc := gracefulConn{Conn: c}
	gc.numCons.inc()

	return gc
}

func (gc gracefulConn) Close() error {
	gc.numCons.dec()

	return gc.Conn.Close()
}
