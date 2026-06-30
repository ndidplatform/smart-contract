/**
 * Copyright (c) 2018, 2019 National Digital ID COMPANY LIMITED
 *
 * This file is part of NDID software.
 *
 * NDID is the free software: you can redistribute it and/or modify it under
 * the terms of the Affero GNU General Public License as published by the
 * Free Software Foundation, either version 3 of the License, or any later
 * version.
 *
 * NDID is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
 * See the Affero GNU General Public License for more details.
 *
 * You should have received a copy of the Affero GNU General Public License
 * along with the NDID source code. If not, see https://www.gnu.org/licenses/agpl.txt.
 *
 * Please contact info@ndid.co.th for any further questions
 *
 */

package app

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

const (
	initialStateDataFilename     string = "data"
	initialStateMetadataFilename string = "metadata"
)

const (
	initialStateDataTypeFile      = "file"
	initialStateDataTypeGoLevelDB = "goleveldb"
)

const (
	syncWriteEvery = 50000
)

const (
	logProgressEvery = 100000
)

var (
	initialStateHashKey = []byte("INITIAL_STATE_HASH")
)

type InitialStateMetadata struct {
	TotalKeyCount    int64  `json:"total_key_count"`
	DataType         string `json:"data_type"`
	InitialStateHash []byte `json:"initial_state_hash"`
}

func (appState *AppState) LoadInitialState(logger *logrus.Entry, initialStateDir string) (hash []byte, err error) {
	startTime := time.Now()

	metadataJSON, err := os.ReadFile(filepath.Join(initialStateDir, initialStateMetadataFilename))
	if err != nil {
		return nil, err
	}

	// read metadata
	var initialStateMetadata InitialStateMetadata
	err = json.Unmarshal(metadataJSON, &initialStateMetadata)
	if err != nil {
		return nil, err
	}

	logger.Infof(
		"Initial state data total key count (from metadata): %d",
		initialStateMetadata.TotalKeyCount,
	)

	logger.Infof(
		"Initial state data type: %s",
		initialStateMetadata.DataType,
	)

	logger.Infof(
		"Initial state hash (from metadata): %x",
		initialStateMetadata.InitialStateHash,
	)

	hashDigest := sha256.New()

	keyCount := int64(0)

	processKV := func(key []byte, value []byte) error {
		isInitialStateHashKey := bytes.Equal(key, initialStateHashKey)

		if !isInitialStateHashKey {
			hashDigest.Write(key)
			hashDigest.Write(actionSet)
			hashDigest.Write(value)
		}

		if keyCount+1%syncWriteEvery == 0 || keyCount+1 == initialStateMetadata.TotalKeyCount {
			err = appState.db.SetSync(key, value)
			if err != nil {
				return err
			}
		} else {
			err = appState.db.Set(key, value)
			if err != nil {
				return err
			}
		}

		if !isInitialStateHashKey {
			keyCount++
		}

		if keyCount%logProgressEvery == 0 {
			logger.Infof(
				"Initial state data keys written: %d/%d (%.2f%%)",
				keyCount,
				initialStateMetadata.TotalKeyCount,
				(float64(keyCount)/float64(initialStateMetadata.TotalKeyCount))*100,
			)
		}

		return nil
	}

	switch initialStateMetadata.DataType {
	case initialStateDataTypeFile:
		dataFile, err := os.Open(filepath.Join(initialStateDir, initialStateDataFilename))
		if err != nil {
			return nil, err
		}
		defer dataFile.Close()

		reader := bufio.NewReader(dataFile)
		for {
			line, err := reader.ReadString('\n')

			if err != nil {
				if err == io.EOF {
					break
				} else {
					logger.Fatalf("initial state data read error at line: %d err: %+v", keyCount+1, err)
					return nil, err
				}
			}

			var kv KeyValue
			err = json.Unmarshal([]byte(line), &kv)
			if err != nil {
				return nil, err
			}

			err = processKV(kv.Key, kv.Value)
			if err != nil {
				return nil, err
			}
		}
	case initialStateDataTypeGoLevelDB:
		db, err := leveldb.OpenFile(
			path.Join(initialStateDir, initialStateDataFilename+".db"),
			&opt.Options{
				ReadOnly: true,
			},
		)
		if err != nil {
			return nil, err
		}

		iter := db.NewIterator(nil, nil)
		defer iter.Release()

		for iter.Next() {
			key := iter.Key()
			value := iter.Value()

			err = processKV(key, value)
			if err != nil {
				return nil, err
			}
		}
		if err := iter.Error(); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported initial state data type: %s", initialStateMetadata.DataType)
	}

	logger.Infof("Initial state data loaded, key count: %d, time used: %s", keyCount, time.Since(startTime))

	hash = hashDigest.Sum(nil)

	logger.Infof("Initial state hash: %x", hash)

	if !bytes.Equal(hash, initialStateMetadata.InitialStateHash) {
		return nil, fmt.Errorf("initial state hash mismatch")
	}

	return hash, nil
}

func (appState *AppState) CheckInitialState(logger *logrus.Entry) (hasInitialState bool, hash []byte, err error) {
	startTime := time.Now()

	initialStateHash, err := appState.db.Get(initialStateHashKey)
	if err != nil {
		return false, nil, err
	}

	if len(initialStateHash) == 0 {
		return false, nil, nil
	}

	logger.Infof(
		"Initial state hash (from state DB): %x",
		initialStateHash,
	)

	skipInitialStateHashVerification := false
	if val := getEnv("ABCI_SKIP_INITIAL_STATE_HASH_VERIFICATION", "false"); val == "true" {
		skipInitialStateHashVerification = true
	}

	if skipInitialStateHashVerification {
		logger.Infof("Initial state hash verification skipped")
		return true, initialStateHash, nil
	}

	logger.Infof("Verifying initial state hash")

	hashDigest := sha256.New()

	keyCount := int64(0)

	iter, err := appState.db.Iterator(nil, nil)
	if err != nil {
		return true, nil, err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		value := iter.Value()

		isInitialStateHashKey := bytes.Equal(key, initialStateHashKey)

		if !isInitialStateHashKey {
			hashDigest.Write(key)
			hashDigest.Write(actionSet)
			hashDigest.Write(value)

			keyCount++
		}

		if keyCount%logProgressEvery == 0 {
			logger.Infof(
				"Initial state data keys read: %d",
				keyCount,
			)
		}
	}
	iter.Close()

	hash = hashDigest.Sum(nil)

	if !bytes.Equal(hash, initialStateHash) {
		return true, hash, fmt.Errorf("initial state hash mismatch")
	}

	logger.Infof("Initial state hash verified, key count: %d, time used: %s", keyCount, time.Since(startTime))

	return true, hash, nil
}
