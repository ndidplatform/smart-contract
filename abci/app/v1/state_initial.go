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
	"crypto/sha256"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	initialStateDataFilename     string = "data"
	initialStateMetadataFilename string = "metadata"
)

const (
	syncWriteEvery = 50000
)

const (
	logProgressEvery = 100000
)

type InitialStateMetadata struct {
	TotalKeyCount int64 `json:"total_key_count"`
}

func (appState *AppState) LoadInitialState(logger *logrus.Entry, initialStateDir string) (hash []byte, err error) {
	startTime := time.Now()

	metadataJSON, err := ioutil.ReadFile(filepath.Join(initialStateDir, initialStateMetadataFilename))
	if err != nil {
		return nil, err
	}

	// read metadata
	var initialStateMetadata InitialStateMetadata
	err = json.Unmarshal(metadataJSON, &initialStateMetadata)
	if err != nil {
		panic(err)
	}

	logger.Infof(
		"Initial state data total key count: %d",
		initialStateMetadata.TotalKeyCount,
	)

	dataFile, err := os.Open(filepath.Join(initialStateDir, initialStateDataFilename))
	if err != nil {
		return nil, err
	}
	defer dataFile.Close()

	hashDigest := sha256.New()

	keyCount := int64(0)

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
			panic(err)
		}

		hashDigest.Write(kv.Key)
		hashDigest.Write(actionSet)
		hashDigest.Write(kv.Value)

		if keyCount+1%syncWriteEvery == 0 || keyCount+1 == initialStateMetadata.TotalKeyCount {
			err = appState.db.SetSync(kv.Key, kv.Value)
			if err != nil {
				return nil, err
			}
		} else {
			err = appState.db.Set(kv.Key, kv.Value)
			if err != nil {
				return nil, err
			}
		}

		keyCount++

		if keyCount%logProgressEvery == 0 {
			logger.Infof(
				"Initial state data keys written: %d/%d (%.2f%%)",
				keyCount,
				initialStateMetadata.TotalKeyCount,
				(float64(keyCount)/float64(initialStateMetadata.TotalKeyCount))*100,
			)
		}
	}

	logger.Infof("Initial state data loaded, time used: %s", time.Since(startTime))

	hash = hashDigest.Sum(nil)

	return hash, nil
}
