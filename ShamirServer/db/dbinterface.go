package db



type Database interface {
	Put(key []byte,value []byte) error
	Get(key[]byte)([]byte, error)
	Delete(key []byte) error
	Close()
}

