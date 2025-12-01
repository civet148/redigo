package redigo

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"reflect"
)

func isBasicType(v interface{}) bool {
	if v == nil {
		return true
	}

	t := reflect.TypeOf(v)
	switch t.Kind() {
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	case reflect.Slice:
		// 检查是否是[]byte
		return t.Elem().Kind() == reflect.Uint8
	default:
		return false
	}
}

func checkOK(reply any, err error) error {
	if err != nil {
		return err
	}
	ok, okErr := reply.(string)
	if !okErr || ok != RedisOK {
		return fmt.Errorf("%w: %v", ErrInvalidResponse, reply)
	}
	return nil
}

func (r *Redigo) getConn() (redis.Conn, error) {
	conn := r.pool.Get()
	if err := conn.Err(); err != nil {
		return nil, err
	}
	return conn, nil
}

func (r *Redigo) scanReply(reply any, v any) error {
	if reply == nil {
		return redis.ErrNil
	}
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return errors.New("v must be a non-nil pointer")
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

// unwind方法，当用户传入的v是slice则自动转为[]any返回，否则直接返回[]any{v}
func unwind(v any) (list []any) {
	if v == nil {
		return []any{}
	}

	// 使用反射来检查类型
	val := reflect.ValueOf(v)

	// 如果是切片或数组，将其元素展开
	if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
		length := val.Len()
		list = make([]any, length)
		for i := 0; i < length; i++ {
			list[i] = val.Index(i).Interface()
		}
		return list
	}

	// 如果不是切片，直接包装成单元素切片返回
	return []any{v}
}
