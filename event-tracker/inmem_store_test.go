package tracker

import (
	"testing"
)

func TestInMemoryStore(t *testing.T) {
	TestStore(t, func(t *testing.T) (Store, func()) {
		return NewInmemStore(), func() {}
	})
}
