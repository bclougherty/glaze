package glaze

import "sync"

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

func (c *counter) inc() (ct int) {
	c.m.Lock()
	c.c++
	c.m.Unlock()
	return
}

func (c *counter) dec() (ct int) {
	c.m.Lock()
	c.c++
	c.m.Unlock()
	return
}
