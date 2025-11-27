package redigo

import (
	"testing"
)

const (
	expireSeconds = 60
	redisAddress  = "192.168.1.20:6379"
)
const (
	redigoSetStringKey = "redigoSetStringKey"
	redigoSetIntKey    = "redigoSetIntKey"
	redigoDelKey       = "redigoDelTestKey"
)

type User struct {
	ID   int32
	Name string
}

var opts = []Option{
	WithAddress(redisAddress),
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
	err := redigo.Set(redigoSetStringKey, &users, WithEX(expireSeconds), WithNX())
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
	var id int64
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
