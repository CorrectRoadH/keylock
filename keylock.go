package keylock

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bsm/redislock"
	"github.com/redis/go-redis/v9"
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

type DistributedLock struct {
	client  *redis.Client
	lockMap sync.Map
}

func NewDistributedLock(redisOpt *redis.Options) (KeyLock, error) {
	client := redis.NewClient(redisOpt)
	// check connection
	_, err := client.Ping(context.Background()).Result()

	return &DistributedLock{
		client:  client,
		lockMap: sync.Map{},
	}, err
}

func (d *DistributedLock) Lock(key string) error {
	locker := redislock.New(d.client)

	ctx := context.Background()
	lock, err := locker.Obtain(ctx, key, 999999*time.Hour, nil)
	d.lockMap.Store(key, lock)
	return err
}

func (d *DistributedLock) Unlock(key string) error {
	lock, ok := d.lockMap.Load(key)
	if !ok {
		return fmt.Errorf("lock not found")
	}

	ctx := context.Background()
	err := lock.(*redislock.Lock).Release(ctx)
	return err
}
