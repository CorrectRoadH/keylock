package tests

import (
	"testing"

	"github.com/CorrectRoadH/keylock"
	"github.com/redis/go-redis/v9"
	"github.com/sourcegraph/conc"
	"github.com/stretchr/testify/assert"
)

func TestSingleThreadCountInDistribute(t *testing.T) {
	var wg conc.WaitGroup

	lock, _ := keylock.NewDistributedLock(&redis.Options{
		Addr: "localhost:6379",
	})
	server := IncServer{
		keylock: lock,
		count:   0,
		t:       t,
	}

	for i := 0; i < 100; i++ {
		wg.Go(server.Inc)
	}
	wg.Wait()

	assert.Equal(t, 100, server.count)
}

func TestMultipleThreadCountInDistribute(t *testing.T) {
	var wg conc.WaitGroup
	var wg2 conc.WaitGroup

	lock, _ := keylock.NewDistributedLock(&redis.Options{
		Addr: "localhost:6379",
	})
	server := IncTwoServer{
		keylock: lock,
		count:   0,
		count2:  0,
		t:       t,
	}

	for i := 0; i < 100; i++ {
		wg.Go(server.Inc1)
	}
	for i := 0; i < 100; i++ {
		wg2.Go(server.Inc2)
	}

	wg.Wait()
	wg2.Wait()

	assert.Equal(t, 100, server.count)
	assert.Equal(t, 100, server.count2)
}
