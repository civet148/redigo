package redigo

import "errors"

var (
	ErrKeyExists             = errors.New("key already exists")
	ErrKeyNotExists          = errors.New("key does not exist")
	ErrInvalidResponse       = errors.New("invalid response from server")
	ErrLockAcquisitionFailed = errors.New("failed to acquire lock")
	ErrLockNotHeld           = errors.New("lock not held by this instance")
)
