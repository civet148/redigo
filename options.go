package redigo

import (
	"crypto/tls"
	"fmt"
	"time"
)

const (
	defaultAddress     = "127.0.0.1:6379"
	defaultPassword    = ""
	defaultDB          = 0
	defaultIdleTimeout = 300 * time.Second
)

type Option func(*redigoOptions)

type redigoOptions struct {
	// redis server address, example: 127.0.0.1:6379
	address string

	// redis server password
	password string

	// redis server db, default: 0
	db int

	// connection timeout
	connTimeout *time.Duration

	// Maximum number of idle connections in the pool.
	maxIdle int

	// Maximum number of connections allocated by the pool at a given time.
	// When zero, there is no limit on the number of connections in the pool.
	maxActive int

	// Close connections after remaining idle for this duration. If the value
	// is zero, then idle connections are not closed. Applications should set
	// the timeout to a value less than the server's timeout.
	idleTimeout time.Duration

	// If Wait is true and the pool is at the MaxActive limit, then Get() waits
	// for a connection to be returned to the pool before returning.
	Wait bool

	// Close connections older than this duration. If the value is zero, then
	// the pool does not close connections based on age.
	MaxConnLifetime time.Duration

	clientName string
	useTLS     bool
	skipVerify bool
	tlsConfig  *tls.Config
}

func WithAddress(address string) Option {
	return func(o *redigoOptions) {
		o.address = address
	}
}

func WithPassword(password string) Option {
	return func(o *redigoOptions) {
		o.password = password
	}
}

func WithDB(db int) Option {
	return func(o *redigoOptions) {
		o.db = db
	}
}

func WithMaxIdle(maxIdle int) Option {
	return func(o *redigoOptions) {
		o.maxIdle = maxIdle
	}
}

func WithMaxActive(maxActive int) Option {
	return func(o *redigoOptions) {
		o.maxActive = maxActive
	}
}

func WithConnTimeout(connTimeout time.Duration) Option {
	return func(o *redigoOptions) {
		o.connTimeout = &connTimeout
	}
}

func WithIdleTimeout(idleTimeout time.Duration) Option {
	return func(o *redigoOptions) {
		o.idleTimeout = idleTimeout
	}
}

func WithClientName(clientName string) Option {
	return func(o *redigoOptions) {
		o.clientName = clientName
	}
}

func WithUseTLS(useTLS bool) Option {
	return func(o *redigoOptions) {
		o.useTLS = useTLS
	}
}

func WithSkipVerify(skipVerify bool) Option {
	return func(o *redigoOptions) {
		o.skipVerify = skipVerify
	}
}

func WithTLSConfig(tlsConfig *tls.Config) Option {
	return func(o *redigoOptions) {
		o.tlsConfig = tlsConfig
	}
}

func WithWait(wait bool) Option {
	return func(o *redigoOptions) {
		o.Wait = wait
	}
}

func WithMaxConnLifetime(maxConnLifetime time.Duration) Option {
	return func(o *redigoOptions) {
		o.MaxConnLifetime = maxConnLifetime
	}
}

func checkParams(o *redigoOptions) error {
	if o.address == "" {
		return fmt.Errorf("empty redis address")
	}
	if o.maxActive < o.maxIdle {
		return fmt.Errorf("connections max active must be greater than or equal to idle")
	}
	return nil
}

/*--------------------------------------------------------------------------------------------------------------------*/

type setOptions struct {
	ex int64 //EX expire in seconds
	nx bool  //NX only do when the key not exist
	xx bool  //XX only do when the key exist
	px int64 //PX expire in milliseconds
}

type SetOption func(*setOptions)

func parseSetOptions(opts ...SetOption) *setOptions {
	options := &setOptions{}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

// WithEX set EX option
func WithEX(expire int64) SetOption {
	return func(o *setOptions) {
		o.ex = expire
	}
}

// WithNX set NX option
func WithNX() SetOption {
	return func(o *setOptions) {
		o.nx = true
	}
}

// WithXX set XX option
func WithXX() SetOption {
	return func(o *setOptions) {
		o.xx = true
	}
}

// WithPX set PX option
func WithPX(px int64) SetOption {
	return func(o *setOptions) {
		o.px = px
	}
}

/*--------------------------------------------------------------------------------------------------------------------*/
type PushOption func(*pushOptions)

type pushOptions struct {
}
