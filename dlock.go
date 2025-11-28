package redigo

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"
)

var (
	ErrLockAcquisitionFailed = errors.New("failed to acquire lock")
	ErrLockNotHeld           = errors.New("lock not held by this instance")
)

// Lock represents a distributed lock
type Lock struct {
	key      string
	value    string
	expiry   int64
	redigo   *Redigo
	acquired bool
}

// randomValue generates a random string for lock value
func randomValue() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// BlockLock acquires a distributed lock in blocking mode
// It returns an unlock function and an error if failed to acquire lock
func (r *Redigo) BlockLock(key string, expiry time.Duration) (func() error, error) {
	value, err := randomValue()
	if err != nil {
		return nil, err
	}

	for {
		acquired, err := r.tryLock(key, value, expiry)
		if err != nil {
			return nil, err
		}
		if acquired {
			lock := &Lock{
				key:      key,
				value:    value,
				expiry:   int64(expiry.Seconds()),
				redigo:   r,
				acquired: true,
			}
			return lock.unlock, nil
		}

		// Wait a bit before retrying
		time.Sleep(100 * time.Millisecond)
	}
}

// TryLock tries to acquire a distributed lock with a timeout
// It returns an unlock function and an error if failed to acquire lock within timeout
func (r *Redigo) TryLock(key string, expiry time.Duration, timeout time.Duration) (func() error, error) {
	value, err := randomValue()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, ErrLockAcquisitionFailed
		default:
			acquired, err := r.tryLock(key, value, expiry)
			if err != nil {
				return nil, err
			}
			if acquired {
				lock := &Lock{
					key:      key,
					value:    value,
					expiry:   int64(expiry.Seconds()),
					redigo:   r,
					acquired: true,
				}
				return lock.unlock, nil
			}

			// Wait a bit before retrying
			time.Sleep(50 * time.Millisecond)
		}
	}
}

// tryLock attempts to acquire the lock once
func (r *Redigo) tryLock(key, value string, expiry time.Duration) (bool, error) {
	conn, err := r.getConn()
	if err != nil {
		return false, err
	}
	defer conn.Close()

	// Using SET with NX and EX options to atomically set the key only if it doesn't exist
	// and automatically expire after the specified time
	result, err := conn.Do("SET", key, value, "NX", "EX", int64(expiry.Seconds()))
	if err != nil {
		return false, err
	}

	// If result is nil, it means the key already existed and we couldn't set it
	if result == nil {
		return false, nil
	}

	return true, nil
}

// unlock releases the distributed lock
func (l *Lock) unlock() error {
	if !l.acquired {
		return ErrLockNotHeld
	}

	conn, err := l.redigo.getConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	// Lua script to check if the lock is still held by us and delete it atomically
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`

	result, err := conn.Do("EVAL", script, 1, l.key, l.value)
	if err != nil {
		return err
	}

	// Mark the lock as released
	l.acquired = false

	// If result is 0, it means the lock was not held by us
	if result == int64(0) {
		return ErrLockNotHeld
	}

	return nil
}
