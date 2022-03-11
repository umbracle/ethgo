package tracker

import (
	"testing"
)

func TestInMemoryStore(t *testing.T) {
	TestStore(t, func(t *testing.T) (Entry, func()) {
		return NewInmemStore(), func() {}
	})
}
