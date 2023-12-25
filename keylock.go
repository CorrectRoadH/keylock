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
	var err error
	var lock *redislock.Lock
	for {
		lock, err = locker.Obtain(ctx, key, 1*time.Hour, nil)
		if err == nil {
			break
		}
		if err == redislock.ErrNotObtained {
			fmt.Println("重试")
			continue
		}
		return err
	}

	d.lockMap.Store(key, lock)
	return err
}

func (d *DistributedLock) Unlock(key string) error {
	lock, ok := d.lockMap.Load(key)
	if !ok {
		return fmt.Errorf("lock not found")
	}

	lockInstance := lock.(*redislock.Lock)
	ctx := context.Background()
	_, err := lockInstance.TTL(ctx)
	if err != nil {
		return err
	}
	err = lockInstance.Release(ctx)
	return err
}
