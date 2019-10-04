package trackerboltdb

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/umbracle/go-web3"
)

func testdb(t *testing.T) (*BoltStore, func()) {
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

func TestGetSet(t *testing.T) {
	store, close := testdb(t)
	defer close()

	k1 := []byte{0x1}
	v1 := []byte{0x1}
	v2 := []byte{0x2}

	res, err := store.Get(k1)
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 0 {
		t.Fatal("expected empty")
	}

	// set the entry
	if err := store.Set(k1, v1); err != nil {
		t.Fatal(err)
	}
	res, err = store.Get(k1)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(res, v1) {
		t.Fatal("bad")
	}

	// update the entry
	if err := store.Set(k1, v2); err != nil {
		t.Fatal(err)
	}
	res, err = store.Get(k1)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(res, v2) {
		t.Fatal("bad")
	}
}

func TestStoreLogs(t *testing.T) {
	store, close := testdb(t)
	defer close()

	indx, err := store.LastIndex()
	if err != nil {
		t.Fatal(err)
	}
	if indx != 0 {
		t.Fatal("index should be zero")
	}

	log := web3.Log{
		BlockNumber: 10,
	}
	if err := store.StoreLog(&log); err != nil {
		t.Fatal(err)
	}

	indx, err = store.LastIndex()
	if err != nil {
		t.Fatal(err)
	}
	if indx != 1 {
		t.Fatal("index should be one")
	}

	var log2 web3.Log
	if err := store.GetLog(0, &log2); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(log, log2) {
		t.Fatal("bad")
	}
}

func TestRemoveLogs(t *testing.T) {
	store, close := testdb(t)
	defer close()

	logs := []*web3.Log{}
	for i := uint64(0); i < 10; i++ {
		logs = append(logs, &web3.Log{
			BlockNumber: i,
		})
	}

	if err := store.StoreLogs(logs); err != nil {
		t.Fatal(err)
	}

	if err := store.RemoveLogs(5); err != nil {
		t.Fatal(err)
	}

	indx, err := store.LastIndex()
	if err != nil {
		t.Fatal(err)
	}
	if indx != 5 {
		t.Fatal("bad")
	}
}
