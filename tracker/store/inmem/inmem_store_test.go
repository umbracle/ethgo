package inmem

import (
	"testing"

	"github.com/Ethernal-Tech/ethgo/tracker/store"
)

func TestInMemoryStore(t *testing.T) {
	store.TestStore(t, func(t *testing.T) (store.Store, func()) {
		return NewInmemStore(), func() {}
	})
}
