package redigo

import (
	"errors"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	expireSeconds = 60
	redisAddress  = "127.0.0.1:6379"
	redisPassword = "123456"
)
const (
	redigoNotFoundKey  = "redigoNotFoundKey"
	redigoSetStringKey = "redigoSetStringKey"
	redigoSetIntKey    = "redigoSetIntKey"
	redigoDelKey       = "redigoDelTestKey"
	redigoListKey      = "redigoListKey"
)

type User struct {
	ID   int32
	Name string
}

var opts = []Option{
	WithAddress(redisAddress),
	WithPassword(redisPassword),
}

func TestRedigo_Set(t *testing.T) {
	users := []*User{
		{
			ID:   10086,
			Name: "dianxin",
		},
		{
			ID:   10010,
			Name: "liantong",
		},
	}
	redigo := NewRedigo(opts...)
	var id string
	err := redigo.Get(redigoNotFoundKey, &id)
	if err != nil {
		if errors.Is(err, redis.ErrNil) {
			t.Logf("key [%s] not found", redigoNotFoundKey)
		} else {
			t.Fatal(err)
		}
	}

	err = redigo.Set(redigoSetStringKey, &users, WithEX(expireSeconds), WithNX())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Set key [%s] ok", redigoSetStringKey)

	var ttl int64
	ttl, err = redigo.TTL(redigoSetStringKey)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("TTL key [%s] return [%v]", redigoSetStringKey, ttl)

	err = redigo.Set(redigoSetIntKey, uint64(592609353484206164), WithEX(expireSeconds))
	if err != nil {
		t.Fatal(err)
	}

	err = redigo.Get(redigoSetIntKey, &id)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("GET key [%v] value [%+v]", redigoSetIntKey, id)
}

func TestRedigo_Del(t *testing.T) {
	// First, set a key to ensure it exists
	redigo := NewRedigo(opts...)
	testValue := "test_value_for_deletion"
	err := redigo.Set(redigoDelKey, testValue, WithEX(expireSeconds))
	if err != nil {
		t.Fatal(err)
	}

	var ttl int64
	ttl, err = redigo.TTL(redigoDelKey)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("key %s TTL %v", redigoDelKey, ttl)
	// Now delete the key
	var ret int64
	ret, err = redigo.Del(redigoDelKey)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("delete return %v", ret)

	// Verify the key is deleted by trying to get it
	var v string
	err = redigo.Get(redigoDelKey, &v)
	if err == nil {
		t.Fatal("Expected error when getting deleted key, but got nil")
	}
	t.Logf("DEL key [%v] successful, get returned error: %v", redigoDelKey, err)
}

func TestRedigo_List(t *testing.T) {
	var err error
	redigo := NewRedigo(opts...)
	_, err = redigo.ListPush(redigoListKey, []string{"1", "3", "5", "7"}, WithUnwind())
	if err != nil {
		t.Fatal(err)
	}

	var list []string
	err = redigo.ListRange(redigoListKey, 0, -1, &list)
	if err != nil {
		t.Fatal(err)
	}
	var n int64
	n, err = redigo.ListLen(redigoListKey)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("list count %v range %+v", n, list)

	list = []string{}
	err = redigo.ListPop(redigoListKey, 1, &list, WithRright())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("list pop %+v", list)
}

func TestUnwind(t *testing.T) {
	// 测试普通值
	assert.Equal(t, []any{42}, unwind(42))
	assert.Equal(t, []any{"hello"}, unwind("hello"))
	assert.Equal(t, []any{3.14}, unwind(3.14))

	// 测试nil
	assert.Equal(t, []any{}, unwind(nil))

	// 测试切片
	assert.Equal(t, []any{1, 2, 3}, unwind([]int{1, 2, 3}))
	assert.Equal(t, []any{"a", "b", "c"}, unwind([]string{"a", "b", "c"}))

	// 测试空切片
	assert.Equal(t, []any{}, unwind([]int{}))
	assert.Equal(t, []any{}, unwind([]string{}))

	// 测试数组
	assert.Equal(t, []any{1, 2, 3}, unwind([3]int{1, 2, 3}))

	// 测试混合类型切片
	mixed := []any{1, "hello", 3.14, true}
	assert.Equal(t, mixed, unwind(mixed))
}
