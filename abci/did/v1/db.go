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

package did

import (
	"encoding/binary"
	"encoding/json"
	"strconv"

	"github.com/golang/protobuf/proto"
	"github.com/ndidplatform/smart-contract/v4/abci/utils"
	"github.com/ndidplatform/smart-contract/v4/protos/data"
	dbm "github.com/tendermint/tendermint/libs/db"
)

// TODO: Refactor as app.DB or simply change function names (e.g. SetStateDB to SetDBState, GetCommittedVersionedStateDB to GetDBCommittedVersionedState)

var (
	appStateMetadataKey = []byte("stateKey")
	// nonceKeyPrefix  = []byte("nonce:")
)

type AppStateMetadata struct {
	Height  int64  `json:"height"`
	AppHash []byte `json:"app_hash"`
}

type AppState struct {
	AppStateMetadata
	db                       dbm.DB
	CurrentBlock             int64
	HashData                 []byte
	uncommittedState         map[string][]byte
	uncommittedVersionsState map[string][]int64
}

func NewAppState(db dbm.DB) (appState AppState) {
	appStateMetadata := loadAppStateMetadata(db)
	// "CurrentBlock" init by BeginBlock
	appState = AppState{
		AppStateMetadata:         appStateMetadata,
		db:                       db,
		HashData:                 make([]byte, 0),
		uncommittedState:         make(map[string][]byte),
		uncommittedVersionsState: make(map[string][]int64),
	}
	return appState
}

func loadAppStateMetadata(db dbm.DB) AppStateMetadata {
	appStateMetadataBytes := db.Get(appStateMetadataKey)
	var appStateMetadata AppStateMetadata
	if len(appStateMetadataBytes) != 0 {
		err := json.Unmarshal(appStateMetadataBytes, &appStateMetadata)
		if err != nil {
			panic(err)
		}
	}
	return appStateMetadata
}

func (appState *AppState) SaveMetadata() {
	appStateMetadataBytes, err := json.Marshal(appState.AppStateMetadata)
	if err != nil {
		panic(err)
	}
	appState.db.Set(appStateMetadataKey, appStateMetadataBytes)
}

func (appState *AppState) Set(key, value []byte) {
	appState.HashData = append(appState.HashData, key...)
	appState.HashData = append(appState.HashData, value...)

	appState.uncommittedState[string(key)] = value
}

func (appState *AppState) SetVersioned(key, value []byte) {
	versionsKeyStr := string(key) + "|versions"
	versionsKey := []byte(versionsKeyStr)

	var versions []int64
	var existInUncommittedState bool
	versions, existInUncommittedState = appState.uncommittedVersionsState[versionsKeyStr]
	if !existInUncommittedState {
		keyVersionsProtobuf := appState.db.Get(versionsKey)
		if keyVersionsProtobuf != nil {
			var keyVersions data.KeyVersions
			err := proto.Unmarshal([]byte(keyVersionsProtobuf), &keyVersions)
			if err != nil {
				panic(err) // Should panic or return err?
			}
			versions = keyVersions.Versions
		}
	}

	if len(versions) == 0 || versions[len(versions)-1] != appState.CurrentBlock {
		appState.HashData = append(appState.HashData, versionsKey...)
		versionBytes := make([]byte, 8)
		for _, version := range versions {
			binary.BigEndian.PutUint64(versionBytes, uint64(version))
			appState.HashData = append(appState.HashData, versionBytes...)
		}

		appState.uncommittedVersionsState[versionsKeyStr] = append(versions, appState.CurrentBlock)
	}

	keyWithVersionStr := string(key) + "|" + strconv.FormatInt(appState.CurrentBlock, 10)

	appState.HashData = append(appState.HashData, key...)
	appState.HashData = append(appState.HashData, value...)

	appState.uncommittedState[keyWithVersionStr] = value
}

func (appState *AppState) Get(key []byte, committed bool) (err error, value []byte) {
	if committed {
		return appState.getCommitted(key)
	} else {
		return appState.get(key)
	}
}

func (appState *AppState) get(key []byte) (err error, value []byte) {
	var existInUncommittedState bool
	value, existInUncommittedState = appState.uncommittedState[string(key)]
	if !existInUncommittedState {
		value = appState.db.Get(key)
	}

	return nil, value
}

func (appState *AppState) getCommitted(key []byte) (err error, value []byte) {
	value = appState.db.Get(key)
	return nil, value
}

func (appState *AppState) GetVersioned(key []byte, height int64, committed bool) (err error, value []byte) {
	if committed {
		return appState.getCommittedVersioned(key, height)
	} else {
		return appState.getVersioned(key, height)
	}
}

func (appState *AppState) getVersioned(key []byte, height int64) (err error, value []byte) {
	versionsKeyStr := string(key) + "|versions"
	versionsKey := []byte(versionsKeyStr)

	var versions []int64
	var existInUncommittedState bool
	versions, existInUncommittedState = appState.uncommittedVersionsState[versionsKeyStr]
	if !existInUncommittedState {
		keyVersionsProtobuf := appState.db.Get(versionsKey)
		if keyVersionsProtobuf != nil {
			var keyVersions data.KeyVersions
			err = proto.Unmarshal([]byte(keyVersionsProtobuf), &keyVersions)
			if err != nil {
				return err, nil
			}
			versions = keyVersions.Versions
		}
	}

	if len(versions) == 0 {
		return nil, nil
	}

	var version int64
	if height <= 0 {
		version = versions[len(versions)-1]
	} else {
		for i := len(versions) - 1; i >= 0; i-- {
			version = versions[i]
			if version <= height {
				break
			}
		}
	}

	keyWithVersionStr := string(key) + "|" + strconv.FormatInt(version, 10)

	if existInUncommittedState {
		value = appState.uncommittedState[keyWithVersionStr]
	} else {
		keyWithVersion := []byte(keyWithVersionStr)
		value = appState.db.Get(keyWithVersion)
	}

	return nil, value
}

func (appState *AppState) getCommittedVersioned(key []byte, height int64) (err error, value []byte) {
	versionsKeyStr := string(key) + "|versions"
	versionsKey := []byte(versionsKeyStr)

	var versions []int64
	keyVersionsProtobuf := appState.db.Get(versionsKey)
	var keyVersions data.KeyVersions
	err = proto.Unmarshal([]byte(keyVersionsProtobuf), &keyVersions)
	if err != nil {
		return err, nil
	}
	versions = keyVersions.Versions

	if len(versions) == 0 {
		return nil, nil
	}

	var version int64
	if height <= 0 {
		version = versions[len(versions)-1]
	} else {
		for i := len(versions) - 1; i >= 0; i-- {
			version = versions[i]
			if version <= height {
				break
			}
		}
	}

	keyWithVersionStr := string(key) + "|" + strconv.FormatInt(version, 10)
	keyWithVersion := []byte(keyWithVersionStr)

	value = appState.db.Get(keyWithVersion)
	return nil, value
}

func (appState *AppState) Has(key []byte, committed bool) bool {
	if committed {
		return appState.hasCommitted(key)
	} else {
		return appState.has(key)
	}
}

func (appState *AppState) has(key []byte) bool {
	_, existInUncommittedState := appState.uncommittedState[string(key)]
	if existInUncommittedState {
		return true
	}
	return appState.db.Has(key)
}

func (appState *AppState) hasCommitted(key []byte) bool {
	return appState.db.Has(key)
}

func (appState *AppState) HasVersioned(key []byte, committed bool) bool {
	if committed {
		return appState.hasCommittedVersioned(key)
	} else {
		return appState.hasVersioned(key)
	}
}

func (appState *AppState) hasVersioned(key []byte) bool {
	versionsKeyStr := string(key) + "|versions"
	versionsKey := []byte(versionsKeyStr)

	_, existInUncommittedState := appState.uncommittedVersionsState[versionsKeyStr]
	if existInUncommittedState {
		return true
	}

	return appState.db.Has(versionsKey)
}

func (appState *AppState) hasCommittedVersioned(key []byte) bool {
	versionsKeyStr := string(key) + "|versions"
	versionsKey := []byte(versionsKeyStr)
	return appState.db.Has(versionsKey)
}

func (appState *AppState) Delete(key []byte) {
	if !appState.has(key) {
		return
	}
	appState.HashData = append(appState.HashData, key...)
	appState.HashData = append(appState.HashData, []byte("delete")...) // Remove or replace with something else?

	appState.uncommittedState[string(key)] = nil
}

func (appState *AppState) DeleteVersioned(key []byte) {
	if !appState.hasVersioned(key) {
		return
	}
	appState.SetVersioned(key, nil)
}

func (appState *AppState) Save() {
	batch := appState.db.NewBatch()
	defer batch.Close()

	for key := range appState.uncommittedState {
		value := appState.uncommittedState[key]
		if value != nil {
			batch.Set([]byte(key), value)
		} else {
			batch.Delete([]byte(key))
		}
	}

	for key := range appState.uncommittedVersionsState {
		versions := appState.uncommittedVersionsState[key]
		var keyVersions data.KeyVersions
		keyVersions.Versions = versions
		value, err := utils.ProtoDeterministicMarshal(&keyVersions)
		if err != nil {
			panic(err) // Should panic or return err?
		}
		batch.Set([]byte(key), value)
	}

	batch.WriteSync()

	appState.uncommittedState = make(map[string][]byte)
	appState.uncommittedVersionsState = make(map[string][]int64)
}
