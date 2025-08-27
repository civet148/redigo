package redigo

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"reflect"
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

func (r *Redigo) GetPool() *redis.Pool {
	return r.pool
}

func (r *Redigo) GetConn() (*redis.Conn, error) {
	conn := r.pool.Get()
	if err := conn.Err(); err != nil {
		return nil, err
	}
	return &conn, nil
}

func (r *Redigo) Get(key string, v any) error {
	conn := r.pool.Get()
	if err := conn.Err(); err != nil {
		return err
	}
	defer conn.Close()

	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return errors.New("v must be a non-nil pointer")
	}

	reply, err := conn.Do("GET", key)
	if err != nil {
		return err
	}
	if reply == nil {
		return redis.ErrNil
	}

	elemVal := val.Elem()
	switch elemVal.Kind() {
	case reflect.String:
		s, err := redis.String(reply, nil)
		if err != nil {
			return err
		}
		elemVal.SetString(s)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := redis.Int64(reply, nil)
		if err != nil {
			return err
		}
		elemVal.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := redis.Uint64(reply, nil)
		if err != nil {
			return err
		}
		elemVal.SetUint(u)
	case reflect.Float32, reflect.Float64:
		f, err := redis.Float64(reply, nil)
		if err != nil {
			return err
		}
		elemVal.SetFloat(f)
	case reflect.Bool:
		b, err := redis.Bool(reply, nil)
		if err != nil {
			return err
		}
		elemVal.SetBool(b)
	case reflect.Slice:
		if elemVal.Type().Elem().Kind() == reflect.Uint8 {
			// 处理[]byte类型
			b, err := redis.Bytes(reply, nil)
			if err != nil {
				return err
			}
			elemVal.SetBytes(b)
		} else {
			// 其他切片类型尝试JSON反序列化
			data, err := redis.Bytes(reply, nil)
			if err != nil {
				return err
			}
			return json.Unmarshal(data, v)
		}
	default:
		// 尝试JSON反序列化到结构体
		data, err := redis.Bytes(reply, nil)
		if err != nil {
			return err
		}
		return json.Unmarshal(data, v)
	}
	return nil
}

func (r *Redigo) Set(key string, v any, opts ...SetOption) error {
	options := parseSetOptions(opts...)

	conn := r.pool.Get()
	if err := conn.Err(); err != nil {
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
	ok, okErr := reply.(string)
	if !okErr || ok != RedisOK {
		return fmt.Errorf("%w: %v", ErrInvalidResponse, reply)
	}
	return nil
}
