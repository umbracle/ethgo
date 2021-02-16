package trackerpostgresql

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/umbracle/go-web3"
	"github.com/umbracle/go-web3/tracker/store"

	// Enable postgres for sqlx
	_ "github.com/lib/pq"
)

var _ store.Store = (*PostgreSQLStore)(nil)

// PostgreSQLStore is a tracker store implementation that uses PostgreSQL as a backend.
type PostgreSQLStore struct {
	db *sqlx.DB
}

// NewPostgreSQLStore creates a new PostgreSQL store
func NewPostgreSQLStore(endpoint string) (*PostgreSQLStore, error) {
	db, err := sql.Open("postgres", endpoint)
	if err != nil {
		return nil, err
	}
	return NewSQLStore(db, "postgres")
}

// NewSQLStore creates a new store with an sql driver
func NewSQLStore(db *sql.DB, driver string) (*PostgreSQLStore, error) {
	sqlxDB := sqlx.NewDb(db, driver)

	// create the kv database if it does not exists
	if _, err := db.Exec(kvSQLSchema); err != nil {
		return nil, err
	}
	return &PostgreSQLStore{db: sqlxDB}, nil
}

// Close implements the store interface
func (p *PostgreSQLStore) Close() error {
	return p.db.Close()
}

// Get implements the store interface
func (p *PostgreSQLStore) Get(k string) (string, error) {
	var out string
	if err := p.db.Get(&out, "SELECT val FROM kv WHERE key=$1", string(k)); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return out, nil
}

// ListPrefix implements the store interface
func (p *PostgreSQLStore) ListPrefix(prefix string) ([]string, error) {
	var out []string
	if err := p.db.Select(&out, "SELECT val FROM kv WHERE key LIKE $1", string(prefix)+"%"); err != nil {
		return nil, err
	}
	return out, nil
}

// Set implements the store interface
func (p *PostgreSQLStore) Set(k, v string) error {
	if _, err := p.db.Exec("INSERT INTO kv (key, val) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET val = $2", k, v); err != nil {
		return err
	}
	return nil
}

// GetEntry implements the store interface
func (p *PostgreSQLStore) GetEntry(hash string) (store.Entry, error) {
	tableName := "logs_" + hash
	if _, err := p.db.Exec(logSQLSchema(tableName)); err != nil {
		return nil, err
	}
	e := &Entry{
		table: tableName,
		db:    p.db,
	}
	return e, nil
}

// Entry is an store.Entry implementation
type Entry struct {
	table string
	db    *sqlx.DB
}

// LastIndex implements the store interface
func (e *Entry) LastIndex() (uint64, error) {
	var index uint64
	if err := e.db.Get(&index, "SELECT indx FROM "+e.table+" ORDER BY indx DESC LIMIT 1"); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return index + 1, nil
}

// StoreLogs implements the store interface
func (e *Entry) StoreLogs(logs []*web3.Log) error {
	lastIndex, err := e.LastIndex()
	if err != nil {
		return err
	}

	tx, err := e.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := "INSERT INTO " + e.table + " (indx, tx_index, tx_hash, block_num, block_hash, address, data, topics) VALUES (:indx, :tx_index, :tx_hash, :block_num, :block_hash, :address, :data, :topics)"

	for indx, log := range logs {
		topics := []string{}
		for _, topic := range log.Topics {
			topics = append(topics, topic.String())
		}
		obj := &logObj{
			Index:     lastIndex + uint64(indx),
			TxIndex:   log.TransactionIndex,
			TxHash:    log.TransactionHash.String(),
			BlockNum:  log.BlockNumber,
			BlockHash: log.BlockHash.String(),
			Address:   log.Address.String(),
			Topics:    strings.Join(topics, ","),
		}
		if log.Data != nil {
			obj.Data = "0x" + hex.EncodeToString(log.Data)
		}

		if _, err := tx.NamedExec(query, obj); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// RemoveLogs implements the store interface
func (e *Entry) RemoveLogs(indx uint64) error {
	if _, err := e.db.Exec("DELETE FROM "+e.table+" WHERE indx >= $1", indx); err != nil {
		return err
	}
	return nil
}

// GetLog implements the store interface
func (e *Entry) GetLog(indx uint64, log *web3.Log) error {
	obj := logObj{}
	if err := e.db.Get(&obj, "SELECT * FROM "+e.table+" WHERE indx=$1", indx); err != nil {
		return err
	}

	log.TransactionIndex = obj.TxIndex
	if err := log.TransactionHash.UnmarshalText([]byte(obj.TxHash)); err != nil {
		return err
	}
	log.BlockNumber = obj.BlockNum
	if err := log.BlockHash.UnmarshalText([]byte(obj.BlockHash)); err != nil {
		return err
	}
	if err := log.Address.UnmarshalText([]byte(obj.Address)); err != nil {
		return err
	}

	if obj.Topics != "" {
		log.Topics = []web3.Hash{}
		for _, item := range strings.Split(obj.Topics, ",") {
			var topic web3.Hash
			if err := topic.UnmarshalText([]byte(item)); err != nil {
				return err
			}
			log.Topics = append(log.Topics, topic)
		}
	} else {
		log.Topics = nil
	}

	if obj.Data != "" {
		if !strings.HasPrefix(obj.Data, "0x") {
			return fmt.Errorf("0x prefix not found in data")
		}
		buf, err := hex.DecodeString(obj.Data[2:])
		if err != nil {
			return err
		}
		log.Data = buf
	} else {
		log.Data = nil
	}

	return nil
}

type logObj struct {
	Index     uint64 `db:"indx"`
	TxIndex   uint64 `db:"tx_index"`
	TxHash    string `db:"tx_hash"`
	BlockNum  uint64 `db:"block_num"`
	BlockHash string `db:"block_hash"`
	Address   string `db:"address"`
	Topics    string `db:"topics"`
	Data      string `db:"data"`
}

var kvSQLSchema = `
CREATE TABLE IF NOT EXISTS kv (
	key text unique,
	val text
);
`

func logSQLSchema(name string) string {
	return `
	CREATE TABLE IF NOT EXISTS ` + name + ` (
		indx 		numeric,
		tx_index 	numeric,
		tx_hash 	text,
		block_num 	numeric,
		block_hash 	text,
		address 	text,
		topics 		text,
		data 		text
	);
	`
}
