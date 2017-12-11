package db

import (
	"bytes"
	"encoding/binary"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
)

var (
	addressShamir = ":shamir"
	versionpre    = []byte("version")
	defaultPath   = "./db/shamirdb"
	test          = "./db/testdb"
)

type shamirdb struct {
	filename string
	db       *leveldb.DB
}

func newShamirdb(path string, version int) (*shamirdb, error) {

	if path == "" {
		path = test
	}
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}
	currentVer := make([]byte, binary.MaxVarintLen64)
	//putvarint return number of writer
	currentVer = currentVer[:binary.PutVarint(currentVer, int64(version))]

	versionvalue, getError := db.Get(versionpre, nil)
	switch getError {
	case leveldb.ErrNotFound:
		if err := db.Put(versionpre, currentVer, nil); err != nil {
			db.Close()
			return nil, err
		}

	case nil:
		if !bytes.Equal(versionvalue, currentVer) {
			db.Close()
			if err = os.RemoveAll(path); err != nil {
				return nil, err
			}
			return newShamirdb(path, version)
		}
	}
	return &shamirdb{
		db:       db,
		filename: path,
	}, nil

}

func (db *shamirdb) close() {

	db.db.Close()

}

func makeKey(address []byte, filed string) []byte {

	return append(address, filed...)

}

func (db *shamirdb) putShamir(address []byte, secret []byte) error {

	if err := db.db.Put(makeKey(address, addressShamir), secret, nil); err != nil {
		return err
	}

	return nil
}

func (db *shamirdb) getShamir(address []byte) ([]byte, error) {
	secrets, err := db.db.Get(makeKey(address, addressShamir), nil)
	if err != nil {
		return nil, err
	}
	return secrets, nil
}
