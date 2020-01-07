package trackerboltdb

import (
	"bytes"
	"encoding/binary"

	"github.com/boltdb/bolt"
	"github.com/umbracle/go-web3"
	"github.com/umbracle/go-web3/tracker/store"
)

var _ store.Store = (*BoltStore)(nil)

var (
	dbLogs = []byte("logs")
	dbConf = []byte("conf")
)

// BoltStore is a tracker store implementation.
type BoltStore struct {
	conn *bolt.DB
}

// New creates a new boltdbstore
func New(path string) (*BoltStore, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}
	store := &BoltStore{
		conn: db,
	}
	if err := store.setupDB(); err != nil {
		store.Close()
		return nil, err
	}
	return store, nil
}

func (b *BoltStore) setupDB() error {
	txn, err := b.conn.Begin(true)
	if err != nil {
		return err
	}
	defer txn.Rollback()

	if _, err := txn.CreateBucketIfNotExists(dbConf); err != nil {
		return err
	}
	return txn.Commit()
}

// Close implements the store interface
func (b *BoltStore) Close() error {
	return b.conn.Close()
}

// Get implements the store interface
func (b *BoltStore) Get(k []byte) ([]byte, error) {
	txn, err := b.conn.Begin(false)
	if err != nil {
		return nil, err
	}
	defer txn.Rollback()

	bucket := txn.Bucket(dbConf)
	val := bucket.Get(k)

	return val, nil
}

// ListPrefix implements the store interface
func (b *BoltStore) ListPrefix(prefix []byte) ([][]byte, error) {
	txn, err := b.conn.Begin(false)
	if err != nil {
		return nil, err
	}
	defer txn.Rollback()

	res := [][]byte{}
	c := txn.Bucket(dbConf).Cursor()
	for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
		res = append(res, v)
	}
	return res, nil
}

// Set implements the store interface
func (b *BoltStore) Set(k, v []byte) error {
	txn, err := b.conn.Begin(true)
	if err != nil {
		return err
	}
	defer txn.Rollback()

	bucket := txn.Bucket(dbConf)
	if err := bucket.Put(k, v); err != nil {
		return err
	}
	return txn.Commit()
}

// GetEntry implements the store interface
func (b *BoltStore) GetEntry(hash string) (store.Entry, error) {
	txn, err := b.conn.Begin(true)
	if err != nil {
		return nil, err
	}
	defer txn.Rollback()

	bucketName := append(dbLogs, []byte(hash)...)
	if _, err := txn.CreateBucketIfNotExists(bucketName); err != nil {
		return nil, err
	}
	if err := txn.Commit(); err != nil {
		return nil, err
	}
	e := &Entry{
		conn:   b.conn,
		bucket: bucketName,
	}
	return e, nil
}

// Entry is an store.Entry implementation
type Entry struct {
	conn   *bolt.DB
	bucket []byte
}

// LastIndex implements the store interface
func (e *Entry) LastIndex() (uint64, error) {
	tx, err := e.conn.Begin(false)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	curs := tx.Bucket(e.bucket).Cursor()
	if last, _ := curs.Last(); last != nil {
		return bytesToUint64(last) + 1, nil
	}
	return 0, nil
}

// StoreLog implements the store interface
func (e *Entry) StoreLog(log *web3.Log) error {
	return e.StoreLogs([]*web3.Log{log})
}

// StoreLogs implements the store interface
func (e *Entry) StoreLogs(logs []*web3.Log) error {
	tx, err := e.conn.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	indx, err := e.LastIndex()
	if err != nil {
		return err
	}

	bucket := tx.Bucket(e.bucket)
	for logIndx, log := range logs {
		key := uint64ToBytes(indx + uint64(logIndx))

		val, err := log.MarshalJSON()
		if err != nil {
			return err
		}
		if err := bucket.Put(key, val); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// RemoveLogs implements the store interface
func (e *Entry) RemoveLogs(indx uint64) error {
	indxKey := uint64ToBytes(indx)

	tx, err := e.conn.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	curs := tx.Bucket(e.bucket).Cursor()
	for k, _ := curs.Seek(indxKey); k != nil; k, _ = curs.Next() {
		if err := curs.Delete(); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetLog implements the store interface
func (e *Entry) GetLog(indx uint64, log *web3.Log) error {
	txn, err := e.conn.Begin(false)
	if err != nil {
		return err
	}
	defer txn.Rollback()

	bucket := txn.Bucket(e.bucket)
	val := bucket.Get(uint64ToBytes(indx))

	if err := log.UnmarshalJSON(val); err != nil {
		return err
	}
	return nil
}

func bytesToUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

func uint64ToBytes(u uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, u)
	return buf
}
