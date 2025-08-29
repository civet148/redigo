package redigo

import (
	"encoding/json"
	"errors"
	"github.com/gomodule/redigo/redis"
	"time"
)

const (
	RedisOK = "OK"
)

var (
	ErrKeyExists       = errors.New("key already exists")
	ErrKeyNotExists    = errors.New("key does not exist")
	ErrInvalidResponse = errors.New("invalid response from server")
)

type Redigo struct {
	pool    *redis.Pool
	options *redigoOptions
}

func NewRedigo(opts ...Option) *Redigo {
	options := newDefaultOptions()
	for _, opt := range opts {
		opt(options)
	}
	pool := &redis.Pool{
		MaxActive:       options.maxActive,
		MaxIdle:         options.maxIdle,
		IdleTimeout:     options.idleTimeout,
		Wait:            options.Wait,
		MaxConnLifetime: options.MaxConnLifetime,
		Dial: func() (redis.Conn, error) {
			dialOptions := []redis.DialOption{
				redis.DialDatabase(options.db),
			}
			if options.password != "" {
				dialOptions = append(dialOptions, redis.DialPassword(options.password))
			}
			if options.connTimeout != nil {
				dialOptions = append(dialOptions, redis.DialConnectTimeout(*options.connTimeout))
			}
			if options.clientName != "" {
				dialOptions = append(dialOptions, redis.DialClientName(options.clientName))
			}
			if options.useTLS {
				dialOptions = append(dialOptions, redis.DialUseTLS(options.useTLS))
			}
			if options.skipVerify {
				dialOptions = append(dialOptions, redis.DialTLSSkipVerify(options.skipVerify))
			}
			if options.tlsConfig != nil {
				dialOptions = append(dialOptions, redis.DialTLSConfig(options.tlsConfig))
			}
			return redis.Dial("tcp", options.address, dialOptions...)
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
	return &Redigo{
		pool:    pool,
		options: options,
	}
}

func newDefaultOptions() *redigoOptions {
	return &redigoOptions{
		address:     defaultAddress,
		password:    defaultPassword,
		db:          defaultDB,
		maxIdle:     5,
		maxActive:   150,
		idleTimeout: defaultIdleTimeout,
		Wait:        true,
	}
}

func (r *Redigo) Do(cmd string, args ...any) (any, error) {
	conn := r.pool.Get()
	if err := conn.Err(); err != nil {
		return nil, err
	}
	defer conn.Close()
	return conn.Do(cmd, args...)
}

func (r *Redigo) Get(key string, v any) error {
	conn, err := r.getConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	reply, err := conn.Do("GET", key)
	if err != nil {
		return err
	}
	if reply == nil {
		return redis.ErrNil
	}
	return r.scanReply(reply, v)
}

func (r *Redigo) Set(key string, v any, opts ...SetOption) error {
	options := parseSetOptions(opts...)

	conn, err := r.getConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	var data any
	if isBasicType(v) {
		data = v
	} else {
		jsonData, err := json.Marshal(v)
		if err != nil {
			return err
		}
		data = string(jsonData)
	}
	var args = []any{
		key,
		data,
	}
	if options.ex != 0 {
		args = append(args, "EX", options.ex)
	} else if options.px != 0 {
		args = append(args, "PX", options.px)
	}
	if options.nx {
		args = append(args, "NX")
	} else if options.xx {
		args = append(args, "XX")
	}
	reply, err := conn.Do("SET", args...)
	if err != nil {
		return err
	}

	// 处理条件性设置选项的返回值
	if options.nx || options.xx {
		if reply == nil {
			if options.nx {
				return ErrKeyExists
			} else {
				return ErrKeyNotExists
			}
		}
	}

	// 检查是否为标准OK响应
	if err = checkOK(reply, err); err != nil {
		return err
	}
	return nil
}

func (r *Redigo) Del(key string) error {
	conn, err := r.getConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	reply, err := conn.Do("DEL", key)
	if err != nil {
		return err
	}

	// 检查是否为标准OK响应
	if err = checkOK(reply, err); err != nil {
		return err
	}
	return nil
}

func (r *Redigo) Exists(key string) (bool, error) {
	conn, err := r.getConn()
	if err != nil {
		return false, err
	}
	defer conn.Close()

	reply, err := conn.Do("EXISTS", key)
	if err != nil {
		return false, err
	}

	// 检查是否为标准OK响应
	if err = checkOK(reply, err); err != nil {
		return false, err
	}
	return true, nil
}

func (r *Redigo) Expire(key string, expiration time.Duration) error {
	conn, err := r.getConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	reply, err := conn.Do("EXPIRE", key, expiration.Seconds())
	if err != nil {
		return err
	}

	// 检查是否为标准OK响应
	if err = checkOK(reply, err); err != nil {
		return err
	}
	return nil
}

// Incr increment the value of a key by 1 if v is nil, otherwise v must be an integer or float number
func (r *Redigo) Incr(key string, v any) (reply any, err error) {
	conn, err := r.getConn()
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	var incr = "INCR"
	var args []any
	if v != nil {
		switch v.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			incr = "INCRBY"
		case float32, float64:
			incr = "INCRBYFLOAT"
		default:
			panic("value must be an integer or float number")
		}
	}
	args = append(args, key)
	if v != nil {
		args = append(args, v)
	}
	return conn.Do(incr, args...)
}

func (r *Redigo) Decr(key string, v ...int64) (reply any, err error) {
	conn, err := r.getConn()
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	var delta = int64(1)
	if len(v) != 0 {
		delta = v[0]
	}
	return conn.Do("DECRBY", key, delta)
}
