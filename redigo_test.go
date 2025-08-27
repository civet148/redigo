package redigo

import (
	"log"
	"testing"
)

const (
	redigoStringKey = "redigoStringKey"
)

type User struct {
	ID   int32
	Name string
}

var opts = []Option{
	WithAddress("192.168.1.20:6379"),
}

func TestRedigo_Set(t *testing.T) {
	user := &User{
		ID:   10086,
		Name: "yuxin001",
	}
	redigo := NewRedigo(opts...)
	err := redigo.Set(redigoStringKey, user, WithEX(60), WithNX())
	if err != nil {
		t.Fatal(err)
	}
}

func TestRedigo_Get(t *testing.T) {
	var v User
	redigo := NewRedigo(opts...)
	err := redigo.Get(redigoStringKey, &v)
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("GET %v value=%v \n", redigoStringKey, v)
}
