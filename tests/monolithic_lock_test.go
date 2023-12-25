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
	t       *testing.T
}

func (s *IncServer) Inc() {
	err := s.keylock.Lock("inc")
	assert.Nil(s.t, err)
	s.count++
	err = s.keylock.Unlock("inc")
	assert.Nil(s.t, err)
}

func TestSingleThreadCount(t *testing.T) {
	var wg conc.WaitGroup

	lock, _ := keylock.New()
	server := IncServer{
		keylock: lock,
		count:   0,
		t:       t,
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
	t       *testing.T
}

func (s *IncTwoServer) Inc1() {
	err := s.keylock.Lock("inc1")
	assert.NoError(s.t, err)
	s.count++
	err = s.keylock.Unlock("inc1")
	assert.NoError(s.t, err)
}

func (s *IncTwoServer) Inc2() {
	err := s.keylock.Lock("inc2")
	assert.NoError(s.t, err)
	s.count2++
	err = s.keylock.Unlock("inc2")
	assert.NoError(s.t, err)
}

func TestMultipleThreadCount(t *testing.T) {
	var wg conc.WaitGroup
	var wg2 conc.WaitGroup

	lock, _ := keylock.New()
	server := IncTwoServer{
		keylock: lock,
		count:   0,
		count2:  0,
		t:       t,
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
