package tracker

import (
	"reflect"
	"testing"

	"github.com/umbracle/ethgo"
)

// SetupDB is a function that creates a backend
type SetupDB func(t *testing.T) (Entry, func())

// TestStore tests a tracker store
func TestStore(t *testing.T, setup SetupDB) {
	testRemoveLogs(t, setup)
	testStoreLogs(t, setup)
}

func testStoreLogs(t *testing.T, setup SetupDB) {
	entry, close := setup(t)
	defer close()

	indx, err := entry.LastIndex()
	if err != nil {
		t.Fatal(err)
	}
	if indx != 0 {
		t.Fatal("index should be zero")
	}

	log := ethgo.Log{
		BlockNumber: 10,
	}
	if err := entry.StoreEvent(&Event{Added: []*ethgo.Log{&log}, Indx: -1}); err != nil {
		t.Fatal(err)
	}

	indx, err = entry.LastIndex()
	if err != nil {
		t.Fatal(err)
	}
	if indx != 1 {
		t.Fatal("index should be one")
	}

	var log2 ethgo.Log
	if err := entry.GetLog(0, &log2); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(log, log2) {
		t.Fatal("bad")
	}
}

func testRemoveLogs(t *testing.T, setup SetupDB) {
	entry, close := setup(t)
	defer close()

	logs := []*ethgo.Log{}
	for i := uint64(0); i < 10; i++ {
		logs = append(logs, &ethgo.Log{
			BlockNumber: i,
		})
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
