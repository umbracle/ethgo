package trackerboltdb

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/umbracle/go-web3/tracker/store"
)

func setupDB(t *testing.T) (store.Store, func()) {
	dir, err := ioutil.TempDir("/tmp", "boltdb-test")
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(dir, "test.db")
	store, err := New(path)
	if err != nil {
		t.Fatal(err)
	}

	close := func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	}
	return store, close
}

func TestBoltDBStore(t *testing.T) {
	store.TestStore(t, setupDB)
}
