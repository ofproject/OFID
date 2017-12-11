package db

import (
//"sync"
//"github.com/ethereum/go-ethereum/common"
)

const (
	version = 1
)

type Table struct {
	//mutex sync.Mutex
	db *shamirdb
}

func NewTable(nodbpath string) (*Table, error) {

	db, err := newShamirdb(nodbpath, version)
	if err != nil {
		return nil, err
	}
	tab := &Table{
		db: db,
	}
	return tab, nil

}

func (table *Table) PutShamir(address []byte, secrets []byte) error {

	if err := table.db.putShamir(address, secrets); err != nil {
		return err
	}
	return nil

}

func (table *Table) GetShamirSecrets(address []byte) ([]byte, error) {

	secrets, err := table.db.getShamir(address)
	if err != nil {
		return nil, err
	}
	return secrets, nil

}

func (t *Table) Close() {
	t.db.close()
}
