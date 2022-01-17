package tracker

import (
	"reflect"
	"testing"

	web3 "github.com/umbracle/go-web3"
)

// SetupDB is a function that creates a backend
type SetupDB func(t *testing.T) (Store, func())

// TestStore tests a tracker store
func TestStore(t *testing.T, setup SetupDB) {
	testMultipleStores(t, setup)
	// testGetSet(t, setup)
	testRemoveLogs(t, setup)
	testStoreLogs(t, setup)
	// testPrefix(t, setup)
}

var (
	filterConfig0 = &FilterConfig{
		Hash: "0",
	}
	filterConfig1 = &FilterConfig{
		Hash: "1",
	}
)

func testMultipleStores(t *testing.T, setup SetupDB) {
	store, close := setup(t)
	defer close()

	entry0, err := store.GetEntry(filterConfig0)
	if err != nil {
		t.Fatal(err)
	}
	log := web3.Log{
		BlockNumber: 10,
	}
	if err := entry0.StoreEvent(&Event{Added: []*web3.Log{&log}}); err != nil {
		t.Fatal(err)
	}

	entry1, err := store.GetEntry(filterConfig1)
	if err != nil {
		t.Fatal(err)
	}
	log = web3.Log{
		BlockNumber: 15,
	}
	if err := entry1.StoreEvent(&Event{Added: []*web3.Log{&log}}); err != nil {
		t.Fatal(err)
	}

	index0, err := entry0.LastIndex()
	if err != nil {
		t.Fatal(err)
	}
	if index0 != 1 {
		t.Fatal("bad")
	}

	index1, err := entry1.LastIndex()
	if err != nil {
		t.Fatal(err)
	}
	if index1 != 1 {
		t.Fatal("bad")
	}
}

/*
func testPrefix(t *testing.T, setup SetupDB) {
	store, close := setup(t)
	defer close()

	v1 := "val1"
	v2 := "val2"
	v3 := "val3"

	// add same prefix values
	if err := store.Set(v1, v1); err != nil {
		t.Fatal(err)
	}
	if err := store.Set(v2, v2); err != nil {
		t.Fatal(err)
	}
	if err := store.Set(v3, v3); err != nil {
		t.Fatal(err)
	}

	// add distinct value
	if err := store.Set("a", "b"); err != nil {
		t.Fatal(err)
	}

	checkPrefix := func(prefix string, expected int) {
		res, err := store.ListPrefix(prefix)
		if err != nil {
			t.Fatal(err)
		}
		if len(res) != expected {
			t.Fatalf("%d values expected for prefix '%s' but %d found", expected, string(prefix), len(res))
		}
	}

	checkPrefix("val", 3)
	checkPrefix("a", 1)
	checkPrefix("b", 0)
}
*/

/*
func testGetSet(t *testing.T, setup SetupDB) {
	store, close := setup(t)
	defer close()

	k1 := "k1"
	v1 := "v1"
	v2 := "v2"

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
	if res != v1 {
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
	if res != v2 {
		t.Fatal("bad")
	}
}
*/

func testStoreLogs(t *testing.T, setup SetupDB) {
	store, close := setup(t)
	defer close()

	entry, err := store.GetEntry(filterConfig1)
	if err != nil {
		t.Fatal(err)
	}

	indx, err := entry.LastIndex()
	if err != nil {
		t.Fatal(err)
	}
	if indx != 0 {
		t.Fatal("index should be zero")
	}

	log := web3.Log{
		BlockNumber: 10,
	}
	if err := entry.StoreEvent(&Event{Added: []*web3.Log{&log}, Indx: -1}); err != nil {
		t.Fatal(err)
	}

	indx, err = entry.LastIndex()
	if err != nil {
		t.Fatal(err)
	}
	if indx != 1 {
		t.Fatal("index should be one")
	}

	var log2 web3.Log
	if err := entry.GetLog(0, &log2); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(log, log2) {
		t.Fatal("bad")
	}

	// retrieve entry again
	entry1, err := store.GetEntry(filterConfig1)
	if err != nil {
		t.Fatal(err)
	}
	indx1, err := entry1.LastIndex()
	if err != nil {
		t.Fatal(err)
	}
	if indx1 != indx {
		t.Fatal("bad last index")
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

	entry, err := store.GetEntry(filterConfig1)
	if err != nil {
		t.Fatal(err)
	}

	if err := entry.StoreEvent(&Event{Added: logs, Indx: -1}); err != nil {
		t.Fatal(err)
	}

	if err := entry.StoreEvent(&Event{Indx: 5}); err != nil {
		t.Fatal(err)
	}

	indx, err := entry.LastIndex()
	if err != nil {
		t.Fatal(err)
	}
	if indx != 5 {
		t.Fatalf("index should be 5 but found %d", indx)
	}

	// add again the values
	if err := entry.StoreEvent(&Event{Added: logs[5:], Indx: -1}); err != nil {
		t.Fatal(err)
	}
	indx, err = entry.LastIndex()
	if err != nil {
		t.Fatal(err)
	}
	if indx != 10 {
		t.Fatalf("index should be 10 but found %d", indx)
	}
}
