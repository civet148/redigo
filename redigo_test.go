package redigo

import (
	"log"
	"testing"
)

const (
	expireSeconds = 3600
	redisAddress  = "192.168.1.20:6379"
)
const (
	redigoSetKey = "redigoSetKey"
	redigoDelKey = "redigoDelTestKey"
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
	err := redigo.Set(redigoSetKey, &users, WithEX(expireSeconds), WithNX())
	if err != nil {
		t.Fatal(err)
	}

	var ttl int64
	ttl, err = redigo.TTL(redigoSetKey)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("key %s TTL %v", redigoSetKey, ttl)
}

func TestRedigo_Get(t *testing.T) {
	var v []User
	redigo := NewRedigo(opts...)
	err := redigo.Get(redigoSetKey, &v)
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("GET key [%v] value %+v \n", redigoSetKey, v)
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
	log.Printf("DEL key [%v] successful, get returned error: %v\n", redigoDelKey, err)
}
