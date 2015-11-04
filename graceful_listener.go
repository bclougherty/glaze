package glaze

import (
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"syscall"
	"time"
)

type counter struct {
	m sync.Mutex
	c int
}

func (c *counter) get() (ct int) {
	c.m.Lock()
	ct = c.c
	c.m.Unlock()
	return
}

var connCount counter

type gracefulConn struct {
	net.Conn
}

func (w gracefulConn) Close() error {
	connCount.m.Lock()
	connCount.c--
	connCount.m.Unlock()

	return w.Conn.Close()
}

type signal struct{}

// GracefulListener wraps a net.Listener and provides the necessary features to gracefully restart without dropping connections
type GracefulListener struct {
	net.Listener
	stop    chan signal
	stopped bool
}

// File is a convenience method for getting a GracefulListener's file descriptor
func (gl GracefulListener) File() *os.File {
	tl := gl.Listener.(*net.TCPListener)
	fl, _ := tl.File()
	return fl
}

// Accept accepts a new connection as net.Listener.Accept, but also tracks connection count so that we know how many open connections there are
func (gl GracefulListener) Accept() (c net.Conn, err error) {
	c, err = gl.Listener.Accept()
	if err != nil {
		return
	}

	// Wrap the returned connection, so that we can observe when
	// it is closed.
	c = gracefulConn{Conn: c}

	// Count it
	connCount.m.Lock()
	connCount.c++
	connCount.m.Unlock()

	return
}

// GracefullyListen listens on the TCP network address addr, either creating or inheriting its socket file based on the value of isChild.
func GracefullyListen(addr string, isChild bool) (*GracefulListener, error) {
	var l net.Listener
	var err error

	if isChild {
		log.Print("Listening to existing file descriptor")
		f := os.NewFile(3, "")
		l, err = net.FileListener(f)
	} else {
		log.Print("Listening on a new file descriptor")
		l, err = net.Listen("tcp", addr)
	}

	if err != nil {
		return nil, err
	}

	listener := &GracefulListener{Listener: l, stop: make(chan signal, 1)}

	// this goroutine monitors the channel. Can't do this in
	// Accept (below) because once it enters listener.Listener.Accept()
	// it blocks. We unblock it by closing the fd it is trying to
	// accept(2) on.
	go func() {
		_ = <-listener.stop
		listener.stopped = true
		listener.Listener.Close()
	}()

	return listener, nil
}

// GracefullyServe accepts incoming connections on the GracefulListener l, creating a new service goroutine for each. The service goroutines read requests and then call handler to reply to them.
func GracefullyServe(l *GracefulListener, handler http.Handler, isChild bool) error {
	server := &http.Server{
		Addr:           l.Addr().String(),
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 16,
	}

	if isChild {
		parent := syscall.Getppid()
		log.Printf("main: Killing parent pid: %v", parent)
		syscall.Kill(parent, syscall.SIGTERM)
	}

	// Start servinc
	log.Printf("Serving on %s", l.Addr().String())
	err := server.Serve(l)

	return err
}
