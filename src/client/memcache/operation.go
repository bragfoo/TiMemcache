package memcache

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func (c *Client) set(rw *bufio.ReadWriter, item *Item) error {
	return c.populateOne(rw, "set", item)
}

func (c *Client) add(rw *bufio.ReadWriter, item *Item) error {
	return c.populateOne(rw, "add", item)
}

func (c *Client) onItem(item *Item, fn func(*Client, *bufio.ReadWriter, *Item) error) error {
	addr, err := c.Selector.PickServer(item.Key)
	if err != nil {
		return err
	}
	cn, err := c.getConn(addr)
	if err != nil {
		return err
	}
	defer cn.condRelease(&err)
	if err = fn(c, cn.rw, item); err != nil {
		return err
	}
	return nil
}

func (c *Client) populateOne(rw *bufio.ReadWriter, verb string, item *Item) error {
	log.Println("populateOne :", verb, item.Key, string(item.Value), len(item.Value), item.Flags, item.Expiration, item.casid)
	if !legalKey(item.Key) {
		return ErrMalformedKey
	}
	var err error
	if verb == "cas" {
		_, err = fmt.Fprintf(rw, "%s %s %d %d %d %d\r\n", verb, item.Key, item.Flags, item.Expiration, len(item.Value), item.casid)
		log.Println("cas :", err)
	} else {
		_, err = fmt.Fprintf(rw, "%s %s %d %d %d\r\n", verb, item.Key, item.Flags, item.Expiration, len(item.Value))
	}
	if err != nil {
		return err
	}
	if _, err = rw.Write(item.Value); err != nil {

		return err
	}
	if _, err := rw.Write(crlf); err != nil {

		return err
	}
	if err := rw.Flush(); err != nil {
		return err
	}
	// line
	line, err := rw.ReadSlice('\n')
	log.Println("line info:", string(line))
	if err != nil {
		log.Println("ReadSlice :", err)
		return err
	}
	log.Println("populateOne :", string(line))
	switch {
	case bytes.Equal(line, resultStored):
		return nil
	case bytes.Equal(line, resultNotStored):
		return ErrNotStored
	case bytes.Equal(line, resultExists):
		return ErrCASConflict
	case bytes.Equal(line, resultNotFound):
		return ErrCacheMiss
	}
	return fmt.Errorf("memcache: unexpected response line from %q: %q", verb, string(line))
}

func (c *Client) withKeyAddr(key string, fn func(net.Addr) error) (err error) {
	if !legalKey(key) {
		return ErrMalformedKey
	}
	addr, err := c.Selector.PickServer(key)
	if err != nil {
		return err
	}
	return fn(addr)
}

func (c *Client) getFromAddr(addr net.Addr, keys []string, cb func(*Item)) error {
	log.Println("getFromAddr :", addr, keys)
	return c.withAddrRw(addr, func(rw *bufio.ReadWriter) error {
		log.Println("withAddrRw :", addr)
		log.Println("join keys :", strings.Join(keys, " "))
		if _, err := fmt.Fprintf(rw, "gets %s\r\n", strings.Join(keys, " ")); err != nil {
			return err
		}
		if err := rw.Flush(); err != nil {
			return err
		}
		// log
		buf := make([]byte, 1024)
		rw.Read(buf)
		log.Println("rw:", string(buf))
		// write
		s := strings.NewReader(string(buf))
		br := bufio.NewReader(s)
		b := bytes.NewBuffer(make([]byte, 1024))
		bw := bufio.NewWriter(b)
		yrw := bufio.NewReadWriter(br, bw)
		yrw.WriteString(string(buf))
		yrw.Flush()
		if err := parseGetResponse(yrw.Reader, cb); err != nil {
			return err
		}
		return nil
	})
}

func (c *Client) withAddrRw(addr net.Addr, fn func(*bufio.ReadWriter) error) (err error) {
	cn, err := c.getConn(addr)
	if err != nil {
		return err
	}
	defer cn.condRelease(&err)
	return fn(cn.rw)
}

// parseGetResponse reads a GET response from r and calls cb for each read and allocated Item
func parseGetResponse(r *bufio.Reader, cb func(*Item)) error {
	// read
	for {
		line, err := r.ReadSlice('\n')
		if err != nil {
			log.Println("ReadSlice:", err)
			return err
		}
		log.Println("line info:", string(line))
		if bytes.Equal(line, resultEnd) {
			return nil
		}
		it := new(Item)
		size, err := scanGetResponseLine(line, it)
		log.Println("size:", size)
		if err != nil {
			log.Println("scanGetResponseLine:", err)
			return err
		}
		it.Value = make([]byte, size+2)
		_, err = io.ReadFull(r, it.Value)
		if err != nil {
			log.Println("ReadFull:", err)
			it.Value = nil
			return err
		}
		log.Println("it.Value:", string(it.Value))
		if !bytes.HasSuffix(it.Value, crlf) {
			it.Value = nil
			return fmt.Errorf("memcache: corrupt get result read")
		}
		it.Value = it.Value[:size]
		cb(it)
	}
}

// scanGetResponseLine populates it and returns the declared size of the item.
// It does not read the bytes of the item.
func scanGetResponseLine(line []byte, it *Item) (size int, err error) {
	log.Println("scanGetResponseLine:", it.Key, it.Flags, size, it.casid)
	pattern := "VALUE %s %d %d %d\r\n"
	dest := []interface{}{&it.Key, &it.Flags, &size, &it.casid}
	if bytes.Count(line, space) == 3 {
		pattern = "VALUE %s %d %d\r\n"
		dest = dest[:3]
	}
	n, err := fmt.Sscanf(string(line), pattern, dest...)
	if err != nil || n != len(dest) {
		return -1, fmt.Errorf("memcache: unexpected line in get response: %q", line)
	}
	return size, nil
}
