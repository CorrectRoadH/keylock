package keylock

import (
	"sync"
)

type KeyLock interface {
	Lock(key string) error
	Unlock(key string) error
}

type KeyLockStruct struct {
	locks sync.Map
}

func New() (KeyLock, error) {
	return &KeyLockStruct{
		locks: sync.Map{},
	}, nil
}

func (k *KeyLockStruct) Lock(key string) error {
	lock, _ := k.locks.LoadOrStore(key, &sync.Mutex{})
	lock.(*sync.Mutex).Lock()
	return nil
}

func (k *KeyLockStruct) Unlock(key string) error {
	lock, _ := k.locks.Load(key)
	lock.(*sync.Mutex).Unlock()
	return nil
}
