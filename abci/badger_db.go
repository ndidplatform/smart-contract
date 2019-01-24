package main

import (
	"bytes"
	"fmt"
	"path/filepath"

	"github.com/dgraph-io/badger"

	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/db"
)

func init() {
	dbCreator := func(name string, dir string) (db.DB, error) {
		return NewGoLevelDB(name, dir)
	}
	db.registerDBCreator(LevelDBBackend, dbCreator, false)
	db.registerDBCreator(GoLevelDBBackend, dbCreator, false)
}

var _ db.DB = (*BadgerDB)(nil)

type BadgerDB struct {
	db *badger.DB
}

func nonNilBytes(bz []byte) []byte {
	if bz == nil {
		return []byte{}
	}
	return bz
}

func NewGoLevelDB(name string, dir string) (*BadgerDB, error) {
	return NewGoLevelDBWithOpts(name, dir, nil)
}

func NewGoLevelDBWithOpts(name string, dir string, o *badger.Options) (*BadgerDB, error) {
	dbPath := filepath.Join(dir, name+".db")
	opts := badger.DefaultOptions
	opts.Dir = dbPath
	opts.ValueDir = dbPath
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	database := &BadgerDB{
		db: db,
	}
	return database, nil
}

// Implements DB.
func (db *BadgerDB) Get(key []byte) []byte {
	key = nonNilBytes(key)

	var val []byte
	var err error
	err = db.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		val, err = item.Value()
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil
		}
		panic(err)
	}
	return val
}

// Implements DB.
func (db *BadgerDB) Has(key []byte) bool {
	return db.Get(key) != nil
}

// Implements DB.
func (db *BadgerDB) Set(key []byte, value []byte) {
	key = nonNilBytes(key)
	value = nonNilBytes(value)

	var err error
	err = db.db.Update(func(txn *badger.Txn) error {
		err = txn.Set(key, value)
		return err
	})

	if err != nil {
		cmn.PanicCrisis(err)
	}
}

// Implements DB.
func (db *BadgerDB) SetSync(key []byte, value []byte) {
	// FIXME: sync?
	key = nonNilBytes(key)
	value = nonNilBytes(value)

	var err error
	err = db.db.Update(func(txn *badger.Txn) error {
		err = txn.Set(key, value)
		return err
	})

	if err != nil {
		cmn.PanicCrisis(err)
	}
}

// Implements DB.
func (db *BadgerDB) Delete(key []byte) {
	key = nonNilBytes(key)

	var err error
	err = db.db.Update(func(txn *badger.Txn) error {
		err = txn.Delete(key)
		return err
	})

	if err != nil {
		cmn.PanicCrisis(err)
	}
}

// Implements DB.
func (db *BadgerDB) DeleteSync(key []byte) {
	key = nonNilBytes(key)

	var err error
	err = db.db.Update(func(txn *badger.Txn) error {
		err = txn.Delete(key)
		return err
	})

	if err != nil {
		cmn.PanicCrisis(err)
	}
}

func (db *BadgerDB) DB() *badger.DB {
	return db.db
}

// Implements DB.
func (db *BadgerDB) Close() {
	db.db.Close()
}

// Implements DB.
func (db *BadgerDB) Print() {
	itr := db.Iterator(nil, nil)
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		key := itr.Key()
		value := itr.Value()
		fmt.Printf("[%X]:\t[%X]\n", key, value)
	}
}

// Implements DB.
func (db *BadgerDB) Stats() map[string]string {
	// TODO
	keys := []string{}

	stats := make(map[string]string)
	// for _, key := range keys {
	// 	str := db.db.PropertyValue(key)
	// 	stats[key] = str
	// }
	return stats
}

//----------------------------------------
// Batch

// Implements DB.
func (db *BadgerDB) NewBatch() db.Batch {
	batch := badger.
	return &badgerDBBatch{db, batch}
}

type badgerDBBatch struct {
	db *BadgerDB
	batch *badger.
}

// Implements Batch.
func (mBatch *badgerDBBatch) Set(key, value []byte) {
	mBatch.batch.Put(key, value)
}

// Implements Batch.
func (mBatch *badgerDBBatch) Delete(key []byte) {
	mBatch.batch.Delete(key)
}

// Implements Batch.
func (mBatch *badgerDBBatch) Write() {
	err := mBatch.db.db.Write(mBatch.batch, &opt.WriteOptions{Sync: false})
	if err != nil {
		panic(err)
	}
}

// Implements Batch.
func (mBatch *badgerDBBatch) WriteSync() {
	err := mBatch.db.db.Write(mBatch.batch, &opt.WriteOptions{Sync: true})
	if err != nil {
		panic(err)
	}
}

//----------------------------------------
// Iterator

// Implements DB.
func (db *BadgerDB) Iterator(start, end []byte) db.Iterator {
	itr := db.db.NewIterator(nil, nil)
	return newBadgerDBIterator(itr, start, end, false)
}

// Implements DB.
func (db *BadgerDB) ReverseIterator(start, end []byte) db.Iterator {
	itr := db.db.NewIterator(nil, nil)
	return newBadgerDBIterator(itr, start, end, true)
}

type badgerDBIterator struct {
	source    badger.Iterator
	start     []byte
	end       []byte
	isReverse bool
	isInvalid bool
}

var _ db.Iterator = (*badgerDBIterator)(nil)

func newBadgerDBIterator(source badger.Iterator, start, end []byte, isReverse bool) *badgerDBIterator {
	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		if isReverse {
			opts.Reverse = true
		}
		source = txn.NewIterator(opts)
		return nil
	})

	if isReverse {
		if end == nil {
			source.Last()
		} else {
			source.Seek(end)
			valid := source.Valid()
			if valid {
				item := source.Item()
				eoakey := item.Key() // end or after key
				if bytes.Compare(end, eoakey) <= 0 {
					source.Rewind()
				}
			} else {
				source.Last()
			}
		}
	} else {
		if start == nil {
			source.First()
		} else {
			source.Seek(start)
		}
	}
	return &badgerDBIterator{
		source:    source,
		start:     start,
		end:       end,
		isReverse: isReverse,
		isInvalid: false,
	}
}

// Implements Iterator.
func (itr *badgerDBIterator) Domain() ([]byte, []byte) {
	return itr.start, itr.end
}

// Implements Iterator.
func (itr *badgerDBIterator) Valid() bool {

	// Once invalid, forever invalid.
	if itr.isInvalid {
		return false
	}

	// Panic on DB error.  No way to recover.
	itr.assertNoError()

	// If source is invalid, invalid.
	if !itr.source.Valid() {
		itr.isInvalid = true
		return false
	}

	// If key is end or past it, invalid.
	var start = itr.start
	var end = itr.end
	var key = itr.source.Key()

	if itr.isReverse {
		if start != nil && bytes.Compare(key, start) < 0 {
			itr.isInvalid = true
			return false
		}
	} else {
		if end != nil && bytes.Compare(end, key) <= 0 {
			itr.isInvalid = true
			return false
		}
	}

	// Valid
	return true
}

// Implements Iterator.
func (itr *badgerDBIterator) Key() []byte {
	// Key returns a copy of the current key.
	// See https://github.com/syndtr/goleveldb/blob/52c212e6c196a1404ea59592d3f1c227c9f034b2/leveldb/iterator/iter.go#L88
	itr.assertNoError()
	itr.assertIsValid()
	return cp(itr.source.Key())
}

// Implements Iterator.
func (itr *badgerDBIterator) Value() []byte {
	// Value returns a copy of the current value.
	// See https://github.com/syndtr/goleveldb/blob/52c212e6c196a1404ea59592d3f1c227c9f034b2/leveldb/iterator/iter.go#L88
	itr.assertNoError()
	itr.assertIsValid()
	return cp(itr.source.Value())
}

// Implements Iterator.
func (itr *badgerDBIterator) Next() {
	itr.assertNoError()
	itr.assertIsValid()
	if itr.isReverse {
		itr.source.Rewind()
	} else {
		itr.source.Next()
	}
}

// Implements Iterator.
func (itr *badgerDBIterator) Close() {
	itr.source.Close()
}

func (itr *badgerDBIterator) assertNoError() {
	if err := itr.source.Error(); err != nil {
		panic(err)
	}
}

func (itr badgerDBIterator) assertIsValid() {
	if !itr.Valid() {
		panic("badgerDBIterator is invalid")
	}
}
