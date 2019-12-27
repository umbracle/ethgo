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
	db, err := sqlx.Connect("postgres", endpoint)
	if err != nil {
		return nil, err
	}

	// create the db (TODO, test)
	if _, err := db.Exec(sqlSchema); err != nil {
		return nil, err
	}

	return &PostgreSQLStore{db: db}, nil
}

// Close implements the store interface
func (p *PostgreSQLStore) Close() error {
	return p.db.Close()
}

// Get implements the store interface
func (p *PostgreSQLStore) Get(k []byte) ([]byte, error) {
	var out string
	if err := p.db.Get(&out, "SELECT val FROM kv WHERE key=$1", string(k)); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return []byte(out), nil
}

// Set implements the store interface
func (p *PostgreSQLStore) Set(k, v []byte) error {
	if _, err := p.db.Exec("INSERT INTO kv (key, val) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET val = $2", string(k), string(v)); err != nil {
		return err
	}
	return nil
}

// LastIndex implements the store interface
func (p *PostgreSQLStore) LastIndex() (uint64, error) {
	var index uint64
	if err := p.db.Get(&index, "SELECT index FROM logs ORDER BY index DESC LIMIT 1"); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return index + 1, nil
}

// StoreLogs implements the store interface
func (p *PostgreSQLStore) StoreLogs(logs []*web3.Log) error {
	lastIndex, err := p.LastIndex()
	if err != nil {
		return err
	}

	tx, err := p.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := "INSERT INTO logs (index, tx_index, tx_hash, block_num, block_hash, address, data, topics) VALUES (:index, :tx_index, :tx_hash, :block_num, :block_hash, :address, :data, :topics)"

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
func (p *PostgreSQLStore) RemoveLogs(indx uint64) error {
	if _, err := p.db.Exec("DELETE FROM logs WHERE index >= $1", indx); err != nil {
		return err
	}
	return nil
}

// GetLog implements the store interface
func (p *PostgreSQLStore) GetLog(indx uint64, log *web3.Log) error {
	obj := logObj{}
	if err := p.db.Get(&obj, "SELECT * FROM logs WHERE index=$1", indx); err != nil {
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
	Index     uint64 `db:"index"`
	TxIndex   uint64 `db:"tx_index"`
	TxHash    string `db:"tx_hash"`
	BlockNum  uint64 `db:"block_num"`
	BlockHash string `db:"block_hash"`
	Address   string `db:"address"`
	Topics    string `db:"topics"`
	Data      string `db:"data"`
}

var sqlSchema = `
CREATE TABLE kv (
	key text unique,
	val text
);

CREATE TABLE logs (
	index 		numeric,
	tx_index 	numeric,
	tx_hash 	text,
	block_num 	numeric,
	block_hash 	text,
	address 	text,
	topics 		text,
	data 		text
);
`
