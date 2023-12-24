package tests

import (
	"testing"

	"github.com/CorrectRoadH/keylock"
	"github.com/sourcegraph/conc"
	"github.com/stretchr/testify/assert"
)

type IncServer struct {
	keylock keylock.KeyLock
	count   int
}

func (s *IncServer) Inc() {
	s.keylock.Lock("inc")
	s.count++
	s.keylock.Unlock("inc")
}

func TestSingleThreadCount(t *testing.T) {
	var wg conc.WaitGroup

	lock, _ := keylock.New()
	server := IncServer{
		keylock: lock,
		count:   0,
	}

	for i := 0; i < 10000; i++ {
		wg.Go(server.Inc)
	}
	wg.Wait()

	assert.Equal(t, 10000, server.count)
}

type IncTwoServer struct {
	keylock keylock.KeyLock
	count   int
	count2  int
}

func (s *IncTwoServer) Inc1() {
	s.keylock.Lock("inc")
	s.count++
	s.keylock.Unlock("inc")
}

func (s *IncTwoServer) Inc2() {
	s.keylock.Lock("inc2")
	s.count2++
	s.keylock.Unlock("inc2")
}

func TestMultipleThreadCount(t *testing.T) {
	var wg conc.WaitGroup
	var wg2 conc.WaitGroup

	lock, _ := keylock.New()
	server := IncTwoServer{
		keylock: lock,
		count:   0,
		count2:  0,
	}

	for i := 0; i < 10000; i++ {
		wg.Go(server.Inc1)
	}
	for i := 0; i < 10000; i++ {
		wg2.Go(server.Inc2)
	}

	wg.Wait()
	wg2.Wait()

	assert.Equal(t, 10000, server.count)
	assert.Equal(t, 10000, server.count2)
}
