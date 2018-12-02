// Package memcache provides a client for the memcached cache server.
package memcache

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"hash/crc32"
	"net"
	"strconv"
	"sync"
	"time"
)

const (
	// DefaultTimeout is the default socket read/write timeout.
	DefaultTimeout = 100 * time.Millisecond
	// DefaultMaxIdleConns is the default maximum number of idle connections kept for any single address.
	DefaultMaxIdleConns = 2
	// arbitrary buffered channel size, for readability
	buffered = 8
)

// ServerList is a simple ServerSelector. Its zero value is usable.
type ServerList struct {
	mu    sync.RWMutex
	addrs []net.Addr
}

// ServerSelector is the interface that selects a memcache server as a function of the item's key.
// All ServerSelector implementations must be safe for concurrent use by multiple goroutines.
type ServerSelector interface {
	// PickServer returns the server address that a given item should be shared onto.
	PickServer(key string) (net.Addr, error)
	Each(func(net.Addr) error) error
}

// conn is a connection to a server.
type conn struct {
	nc   net.Conn
	rw   *bufio.ReadWriter
	addr net.Addr
	c    *Client
}

// Client is a memcache client. It is safe for unlocked use by multiple concurrent goroutines.
type Client struct {
	// Timeout specifies the socket read/write timeout. If zero, DefaultTimeout is used.
	Timeout time.Duration
	// MaxIdleConns specifies the maximum number of idle connections that will be maintained per address. If less than one, DefaultMaxIdleConns will be used.
	// Consider your expected traffic rates and latency carefully. This should be set to a number higher than your peak parallel requests.
	MaxIdleConns int
	Selector     ServerSelector
	Freeconn     map[string][]*conn
	lk           sync.Mutex
}

// ConnectTimeoutError is the error type used when it takes
// too long to connect to the desired host. This level of
// detail can generally be ignored.
type ConnectTimeoutError struct {
	Addr net.Addr
}

// staticAddr caches the Network() and String() values from any net.Addr.
type staticAddr struct {
	ntw, str string
}

// keyBufPool returns []byte buffers for use by PickServer's call to crc32.ChecksumIEEE to avoid allocations.
// But doesn't avoid the copies, which at least are bounded in size and small
var keyBufPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 256)
		return &b
	},
}

// PickServer is pick server
func (ss *ServerList) PickServer(key string) (net.Addr, error) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	if len(ss.addrs) == 0 {
		return nil, ErrNoServers
	}
	if len(ss.addrs) == 1 {
		return ss.addrs[0], nil
	}
	bufp := keyBufPool.Get().(*[]byte)
	n := copy(*bufp, key)
	cs := crc32.ChecksumIEEE((*bufp)[:n])
	keyBufPool.Put(bufp)
	return ss.addrs[cs%uint32(len(ss.addrs))], nil
}

// Each server is given equal weight. A server is given more weight if it's listed multiple times.
// Each iterates over each server calling the given function
func (ss *ServerList) Each(f func(net.Addr) error) error {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	for _, a := range ss.addrs {
		if err := f(a); nil != err {
			return err
		}
	}
	return nil
}

func (s *staticAddr) Network() string {
	return s.ntw
}

func (s *staticAddr) String() string {
	return s.str
}

func newStaticAddr(a net.Addr) net.Addr {
	return &staticAddr{
		ntw: a.Network(),
		str: a.String(),
	}
}

// SetServers changes a ServerList's set of servers at runtime and is safe for concurrent use by multiple goroutines.
// SetServers returns an error if any of the server names fail to resolve. No attempt is made to connect to the server.
// If any error is returned, no changes are made to the ServerList.
func (ss *ServerList) SetServers(servers ...string) error {
	naddr := make([]net.Addr, len(servers))
	for i, server := range servers {
		tcpaddr, err := net.ResolveTCPAddr("tcp", server)
		if err != nil {
			return err
		}
		naddr[i] = newStaticAddr(tcpaddr)
	}
	ss.mu.Lock()
	defer ss.mu.Unlock()
	fmt.Println("ss.addrs", naddr)
	ss.addrs = naddr
	return nil
}

// NewFromSelector returns a new Client using the provided ServerSelector.
func NewFromSelector(ss ServerSelector) *Client {
	return &Client{Selector: ss}
}

// New returns a memcache client using the provided server(s). If a server is listed multiple times, it gets a proportional amount of weight.
func New(server ...string) *Client {
	ss := new(ServerList)
	ss.SetServers(server...)
	return NewFromSelector(ss)
}

func (c *Client) netTimeout() time.Duration {
	if c.Timeout != 0 {
		return c.Timeout
	}
	return DefaultTimeout
}

func (cn *conn) extendDeadline() {
	cn.nc.SetDeadline(time.Now().Add(cn.c.netTimeout()))
}

func (c *Client) getFreeConn(addr net.Addr) (cn *conn, ok bool) {
	c.lk.Lock()
	defer c.lk.Unlock()
	if c.Freeconn == nil {
		return nil, false
	}
	freelist, ok := c.Freeconn[addr.String()]
	if !ok || len(freelist) == 0 {
		return nil, false
	}
	cn = freelist[len(freelist)-1]
	c.Freeconn[addr.String()] = freelist[:len(freelist)-1]
	return cn, true
}

func (c *Client) dial(addr net.Addr) (net.Conn, error) {
	type connError struct {
		cn  net.Conn
		err error
	}
	nc, err := net.DialTimeout(addr.Network(), addr.String(), c.netTimeout())
	if err == nil {
		return nc, nil
	}
	if ne, ok := err.(net.Error); ok && ne.Timeout() {
		return nil, &ConnectTimeoutError{addr}
	}
	return nil, err
}

func (c *Client) getConn(addr net.Addr) (*conn, error) {
	cn, ok := c.getFreeConn(addr)
	if ok {
		cn.extendDeadline()
		return cn, nil
	}
	nc, err := c.dial(addr)
	if err != nil {
		return nil, err
	}
	cn = &conn{
		nc:   nc,
		addr: addr,
		rw:   bufio.NewReadWriter(bufio.NewReader(nc), bufio.NewWriter(nc)),
		c:    c,
	}
	cn.extendDeadline()
	return cn, nil
}

// resumableError returns true if err is only a protocol-level cache error.
// This is used to determine whether or not a server connection should be re-used or not.
// If an error occurs, by default we don't reuse the connection, unless it was just a cache error.
func resumableError(err error) bool {
	switch err {
	case ErrCacheMiss, ErrCASConflict, ErrNotStored, ErrMalformedKey:
		return true
	}
	return false
}

func (c *Client) maxIdleConns() int {
	if c.MaxIdleConns > 0 {
		return c.MaxIdleConns
	}
	return DefaultMaxIdleConns
}

func (c *Client) putFreeConn(addr net.Addr, cn *conn) {
	c.lk.Lock()
	defer c.lk.Unlock()
	if c.Freeconn == nil {
		c.Freeconn = make(map[string][]*conn)
	}
	freelist := c.Freeconn[addr.String()]
	if len(freelist) >= c.maxIdleConns() {
		cn.nc.Close()
		return
	}
	c.Freeconn[addr.String()] = append(freelist, cn)
}

// release returns this connection back to the client's free pool
func (cn *conn) release() {
	cn.c.putFreeConn(cn.addr, cn)
}

// condRelease releases this connection if the error pointed to by err is nil (not an error) or is only a protocol level error (e.g. a cache miss).
// The purpose is to not recycle TCP connections that are bad.
func (cn *conn) condRelease(err *error) {
	if *err == nil || resumableError(*err) {
		cn.release()
	} else {
		cn.nc.Close()
	}
}

// Replace writes the given item, but only if the server *does* already hold data for this key
func (c *Client) Replace(item *Item) error {
	return c.onItem(item, (*Client).replace)
}

func (c *Client) replace(rw *bufio.ReadWriter, item *Item) error {
	return c.populateOne(rw, "replace", item)
}

// GetMulti is a batch version of Get.
// The returned map from keys to items may have fewer elements than the input slice, due to memcache cache misses. Each key must be at most 250 bytes in length.
// If no error is returned, the returned map will also be non-nil.
func (c *Client) GetMulti(keys []string) (map[string]*Item, error) {
	var lk sync.Mutex
	m := make(map[string]*Item)
	addItemToMap := func(it *Item) {
		lk.Lock()
		defer lk.Unlock()
		m[it.Key] = it
	}
	keyMap := make(map[net.Addr][]string)
	for _, key := range keys {
		if !legalKey(key) {
			return nil, ErrMalformedKey
		}
		addr, err := c.Selector.PickServer(key)
		if err != nil {
			return nil, err
		}
		keyMap[addr] = append(keyMap[addr], key)
	}
	ch := make(chan error, buffered)
	for addr, keys := range keyMap {
		go func(addr net.Addr, keys []string) {
			ch <- c.getFromAddr(addr, keys, addItemToMap)
		}(addr, keys)
	}
	var err error
	for range keyMap {
		if ge := <-ch; ge != nil {
			err = ge
		}
	}
	return m, err
}

// Touch updates the expiry for the given key.
// The seconds parameter is either a Unix timestamp or, if seconds is less than 1 month, the number of seconds into the future at which time the item will expire.
// Zero means the item has no expiration time. ErrCacheMiss is returned if the key is not in the cache.
// The key must be at most 250 bytes in length.
func (c *Client) Touch(key string, seconds int32) (err error) {
	return c.withKeyAddr(key, func(addr net.Addr) error {
		return c.touchFromAddr(addr, []string{key}, seconds)
	})
}

func (c *Client) touchFromAddr(addr net.Addr, keys []string, expiration int32) error {
	return c.withAddrRw(addr, func(rw *bufio.ReadWriter) error {
		for _, key := range keys {
			if _, err := fmt.Fprintf(rw, "touch %s %d\r\n", key, expiration); err != nil {
				return err
			}
			if err := rw.Flush(); err != nil {
				return err
			}
			line, err := rw.ReadSlice('\n')
			if err != nil {
				return err
			}
			switch {
			case bytes.Equal(line, resultTouched):
				break
			case bytes.Equal(line, resultNotFound):
				return ErrCacheMiss
			default:
				return fmt.Errorf("memcache: unexpected response line from touch: %q", string(line))
			}
		}
		return nil
	})
}

// Increment atomically increments key by delta.
// The return value is the new value after being incremented or an error.
// If the value didn't exist in memcached the error is ErrCacheMiss.
// The value in memcached must be an decimal number, or an error will be returned.
// On 64-bit overflow, the new value wraps around.
func (c *Client) Increment(key string, delta uint64) (newValue uint64, err error) {
	return c.incrDecr("incr", key, delta)
}

func (c *Client) incrDecr(verb, key string, delta uint64) (uint64, error) {
	var val uint64
	err := c.withKeyRw(key, func(rw *bufio.ReadWriter) error {
		line, err := writeReadLine(rw, "%s %s %d\r\n", verb, key, delta)
		if err != nil {
			return err
		}
		switch {
		case bytes.Equal(line, resultNotFound):
			return ErrCacheMiss
		case bytes.HasPrefix(line, resultClientErrorPrefix):
			errMsg := line[len(resultClientErrorPrefix) : len(line)-2]
			return errors.New("memcache: client error: " + string(errMsg))
		}
		val, err = strconv.ParseUint(string(line[:len(line)-2]), 10, 64)
		if err != nil {
			return err
		}
		return nil
	})
	return val, err
}

// Decrement atomically decrements key by delta.
// The return value is the new value after being decremented or an error.
// If the value didn't exist in memcached the error is ErrCacheMiss.
// The value in memcached must be an decimal number, or an error will be returned.
// On underflow, the new value is capped at zero and does not wrap around.
func (c *Client) Decrement(key string, delta uint64) (newValue uint64, err error) {
	return c.incrDecr("decr", key, delta)
}

// Delete deletes the item with the provided key.
// The error ErrCacheMiss is returned if the item didn't already exist in the cache.
func (c *Client) Delete(key string) error {
	return c.withKeyRw(key, func(rw *bufio.ReadWriter) error {
		return writeExpectf(rw, resultDeleted, "delete %s\r\n", key)
	})
}

// DeleteAll deletes all items in the cache.
func (c *Client) DeleteAll() error {
	return c.withKeyRw("", func(rw *bufio.ReadWriter) error {
		return writeExpectf(rw, resultDeleted, "flush_all\r\n")
	})
}

func (c *Client) withKeyRw(key string, fn func(*bufio.ReadWriter) error) error {
	return c.withKeyAddr(key, func(addr net.Addr) error {
		return c.withAddrRw(addr, fn)
	})
}

func writeExpectf(rw *bufio.ReadWriter, expect []byte, format string, args ...interface{}) error {
	line, err := writeReadLine(rw, format, args...)
	if err != nil {
		return err
	}
	switch {
	case bytes.Equal(line, resultOK):
		return nil
	case bytes.Equal(line, expect):
		return nil
	case bytes.Equal(line, resultNotStored):
		return ErrNotStored
	case bytes.Equal(line, resultExists):
		return ErrCASConflict
	case bytes.Equal(line, resultNotFound):
		return ErrCacheMiss
	}
	return fmt.Errorf("memcache: unexpected response line: %q", string(line))
}

// FlushAll is flush all.
func (c *Client) FlushAll() error {
	return c.Selector.Each(c.flushAllFromAddr)
}

// flushAllFromAddr send the flush_all command to the given addr
func (c *Client) flushAllFromAddr(addr net.Addr) error {
	return c.withAddrRw(addr, func(rw *bufio.ReadWriter) error {
		if _, err := fmt.Fprintf(rw, "flush_all\r\n"); err != nil {
			return err
		}
		if err := rw.Flush(); err != nil {
			return err
		}
		line, err := rw.ReadSlice('\n')
		if err != nil {
			return err
		}
		switch {
		case bytes.Equal(line, resultOk):
			break
		default:
			return fmt.Errorf("memcache: unexpected response line from flush_all: %q", string(line))
		}
		return nil
	})
}

func legalKey(key string) bool {
	if len(key) > 250 {
		return false
	}
	for i := 0; i < len(key); i++ {
		if key[i] <= ' ' || key[i] == 0x7f {
			return false
		}
	}
	return true
}

// CompareAndSwap writes the given item that was previously returned by Get, if the value was neither modified or evicted between the
// Get and the CompareAndSwap calls.
// The item's Key should not change between calls but all other item fields may differ.
// ErrCASConflict is returned if the value was modified in between the calls.
// ErrNotStored is returned if the value was evicted in between the calls.
func (c *Client) CompareAndSwap(item *Item) error {
	return c.onItem(item, (*Client).cas)
}

func (c *Client) cas(rw *bufio.ReadWriter, item *Item) error {
	return c.populateOne(rw, "cas", item)
}

func writeReadLine(rw *bufio.ReadWriter, format string, args ...interface{}) ([]byte, error) {
	_, err := fmt.Fprintf(rw, format, args...)
	if err != nil {
		return nil, err
	}
	if err := rw.Flush(); err != nil {
		return nil, err
	}
	line, err := rw.ReadSlice('\n')
	return line, err
}

func (cte *ConnectTimeoutError) Error() string {
	return "memcache: connect timeout to " + cte.Addr.String()
}

// Set func
// writes the given item unconditionally
func (c *Client) Set(item *Item) error {
	return c.onItem(item, (*Client).set)
}

// Add func
// writes the given item, if no value already exists for its key
// ErrNotStored is returned if that condition is not met
func (c *Client) Add(item *Item) error {
	return c.onItem(item, (*Client).add)
}

// Get func
// gets the item for the given key
// ErrCacheMiss is returned for a memcache cache miss
func (c *Client) Get(key string) (item *Item, err error) {
	err = c.withKeyAddr(key, func(addr net.Addr) error {
		return c.getFromAddr(addr, []string{key}, func(it *Item) { item = it })
	})
	if err == nil && item == nil {
		err = ErrCacheMiss
	}
	return
}
