package store

import (
	"bytes"
	"reflect"
	"testing"

	web3 "github.com/umbracle/go-web3"
)

// SetupDB is a function that creates a backend
type SetupDB func(t *testing.T) (Store, func())

// TestStore tests a tracker store
func TestStore(t *testing.T, setup SetupDB) {
	testGetSet(t, setup)
	testRemoveLogs(t, setup)
	testStoreLogs(t, setup)
}

func testGetSet(t *testing.T, setup SetupDB) {
	store, close := setup(t)
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

func testStoreLogs(t *testing.T, setup SetupDB) {
	store, close := setup(t)
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
	if err := store.StoreLogs([]*web3.Log{&log}); err != nil {
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

func testRemoveLogs(t *testing.T, setup SetupDB) {
	store, close := setup(t)
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

	// add again the values
	if err := store.StoreLogs(logs[5:]); err != nil {
		t.Fatal(err)
	}
	indx, err = store.LastIndex()
	if err != nil {
		t.Fatal(err)
	}
	if indx != 10 {
		t.Fatal("bad")
	}
}
