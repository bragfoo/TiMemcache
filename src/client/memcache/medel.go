package memcache

import "errors"

var (
	crlf  = []byte("\r\n")
	space = []byte(" ")

	resultOK                = []byte("OK\r\n")
	resultStored            = []byte("STORED\r\n")
	resultNotStored         = []byte("NOT_STORED\r\n")
	resultExists            = []byte("EXISTS\r\n")
	resultNotFound          = []byte("NOT_FOUND\r\n")
	resultDeleted           = []byte("DELETED\r\n")
	resultEnd               = []byte("END\r\n")
	resultOk                = []byte("OK\r\n")
	resultTouched           = []byte("TOUCHED\r\n")
	resultClientErrorPrefix = []byte("CLIENT_ERROR ")

	// ErrNotStored means that a conditional write operation failed because the condition was not satisfied.
	ErrNotStored = errors.New("memcache: item not stored")
	// ErrCASConflict means that a CompareAndSwap call failed due to the cached value being modified between the Get and the CompareAndSwap.
	// If the cached value was simply evicted rather than replaced, ErrNotStored will be returned instead.
	ErrCASConflict = errors.New("memcache: compare-and-swap conflict")
	// ErrCacheMiss means that a Get failed because the item wasn't present.
	ErrCacheMiss = errors.New("memcache: cache miss")
	// ErrNoServers is returned when no servers are configured or available.
	ErrNoServers = errors.New("memcache: no servers configured or available")
	// ErrServerError means that a server error occurred.
	ErrServerError = errors.New("memcache: server error")
	// ErrMalformedKey is returned when an invalid key is used.
	// Keys must be at maximum 250 bytes long and not contain whitespace or control characters.
	ErrMalformedKey = errors.New("malformed: key is too long or contains invalid characters")
	// ErrNoStats means that no statistics were available.
	ErrNoStats = errors.New("memcache: no statistics available")
)

// Item is an item to be got or stored in a memcached server.
type Item struct {
	// Key is the Item's key (250 bytes maximum).
	Key string

	// Value is the Item's value.
	Value []byte

	// Flags are server-opaque flags whose semantics are entirely
	// up to the app.
	Flags uint32

	// Expiration is the cache expiration time, in seconds: either a relative
	// time from now (up to 1 month), or an absolute Unix epoch time.
	// Zero means the Item has no expiration time.
	Expiration int32

	// Compare and swap ID.
	casid uint64
}
