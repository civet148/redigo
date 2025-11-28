package redigo

import (
	"testing"
	"time"
)

const (
	testLockKey = "test_dlock_key"
)

func TestRedigo_BlockLock(t *testing.T) {
	redigo := NewRedigo(opts...)

	// Acquire lock
	unlock, err := redigo.BlockLock(testLockKey, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	// Release lock
	err = unlock()
	if err != nil {
		t.Fatal(err)
	}
}

func TestRedigo_TryLock(t *testing.T) {
	redigo := NewRedigo(opts...)

	// Acquire lock
	unlock, err := redigo.TryLock(testLockKey, 10*time.Second, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	// Release lock
	err = unlock()
	if err != nil {
		t.Fatal(err)
	}
}

func TestRedigo_TryLockTimeout(t *testing.T) {
	redigo1 := NewRedigo(opts...)
	redigo2 := NewRedigo(opts...)

	// Acquire lock with first client
	unlock1, err := redigo1.BlockLock(testLockKey, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	// Try to acquire the same lock with second client - should timeout
	_, err = redigo2.TryLock(testLockKey, 10*time.Second, 1*time.Second)
	if err == nil {
		t.Fatal("Expected timeout error, but got nil")
	}

	// Release the first lock
	err = unlock1()
	if err != nil {
		t.Fatal(err)
	}

	// Now try to acquire the lock again - should succeed
	unlock2, err := redigo2.TryLock(testLockKey, 10*time.Second, 1*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	// Release the second lock
	err = unlock2()
	if err != nil {
		t.Fatal(err)
	}
}
