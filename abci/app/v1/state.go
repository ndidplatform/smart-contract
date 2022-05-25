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
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"hash"
	"strconv"

	dbm "github.com/tendermint/tm-db"
	"google.golang.org/protobuf/proto"

	"github.com/ndidplatform/smart-contract/v7/abci/utils"
	data "github.com/ndidplatform/smart-contract/v7/protos/data"
)

var (
	appStateMetadataKey = []byte("stateKey")
	// nonceKeyPrefix  = []byte("nonce:")
)

var (
	actionSet    = []byte("SET")
	actionDelete = []byte("DELETE")
)

type AppStateMetadata struct {
	Height  int64  `json:"height"`
	AppHash []byte `json:"app_hash"`
}

type AppState struct {
	AppStateMetadata
	db                       dbm.DB
	CurrentBlockHeight       int64
	HasHashData              bool
	HashDigest               hash.Hash
	uncommittedState         map[string][]byte
	uncommittedVersionsState map[string][]int64
}

func NewAppState(db dbm.DB) (appState *AppState, err error) {
	appStateMetadata, err := loadAppStateMetadata(db)
	if err != nil {
		return nil, err
	}
	appState = &AppState{
		AppStateMetadata:         *appStateMetadata,
		db:                       db,
		CurrentBlockHeight:       appStateMetadata.Height,
		HasHashData:              false,
		HashDigest:               sha256.New(),
		uncommittedState:         make(map[string][]byte),
		uncommittedVersionsState: make(map[string][]int64),
	}
	return appState, nil
}

func loadAppStateMetadata(db dbm.DB) (*AppStateMetadata, error) {
	appStateMetadataBytes, err := db.Get(appStateMetadataKey)
	if err != nil {
		return nil, err
	}
	appStateMetadata := &AppStateMetadata{}
	if len(appStateMetadataBytes) != 0 {
		err := json.Unmarshal(appStateMetadataBytes, &appStateMetadata)
		if err != nil {
			return nil, err
		}
	}
	return appStateMetadata, nil
}

func (appState *AppState) SaveMetadata() error {
	appStateMetadataBytes, err := json.Marshal(appState.AppStateMetadata)
	if err != nil {
		return err
	}
	appState.db.Set(appStateMetadataKey, appStateMetadataBytes)

	return nil
}

// Set value `nil` equals Delete
func (appState *AppState) Set(key, value []byte) {
	appState.HasHashData = true
	appState.HashDigest.Write(key)
	appState.HashDigest.Write(actionSet)
	appState.HashDigest.Write(value)

	appState.uncommittedState[string(key)] = value
}

func (appState *AppState) SetVersioned(key, value []byte) error {
	versionsKeyStr := string(key) + "|versions"
	versionsKey := []byte(versionsKeyStr)

	var versions []int64
	var existInUncommittedState bool
	versions, existInUncommittedState = appState.uncommittedVersionsState[versionsKeyStr]
	if !existInUncommittedState {
		keyVersionsProtobuf, err := appState.db.Get(versionsKey)
		if err != nil {
			return err
		}
		if keyVersionsProtobuf != nil {
			var keyVersions data.KeyVersions
			if err := proto.Unmarshal([]byte(keyVersionsProtobuf), &keyVersions); err != nil {
				return err
			}
			versions = keyVersions.Versions
		}
	}

	if len(versions) == 0 || versions[len(versions)-1] != appState.CurrentBlockHeight {
		appState.HasHashData = true
		appState.HashDigest.Write(versionsKey)
		appState.HashDigest.Write(actionSet)

		versionBytes := make([]byte, 8)
		for _, version := range versions {
			binary.BigEndian.PutUint64(versionBytes, uint64(version))

			appState.HashDigest.Write(versionBytes)
		}

		appState.uncommittedVersionsState[versionsKeyStr] = append(versions, appState.CurrentBlockHeight)
	}

	keyWithVersionStr := string(key) + "|" + strconv.FormatInt(appState.CurrentBlockHeight, 10)

	appState.HasHashData = true
	appState.HashDigest.Write(key)
	appState.HashDigest.Write(actionSet)
	appState.HashDigest.Write(value)

	appState.uncommittedState[keyWithVersionStr] = value

	return nil
}

func (appState *AppState) Get(key []byte, committed bool) (value []byte, err error) {
	if committed {
		return appState.getCommitted(key)
	} else {
		return appState.get(key)
	}
}

func (appState *AppState) get(key []byte) (value []byte, err error) {
	var existInUncommittedState bool
	value, existInUncommittedState = appState.uncommittedState[string(key)]
	if !existInUncommittedState {
		value, err = appState.db.Get(key)
		if err != nil {
			return nil, err
		}
	}

	return value, nil
}

func (appState *AppState) getCommitted(key []byte) (value []byte, err error) {
	value, err = appState.db.Get(key)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (appState *AppState) GetVersioned(key []byte, height int64, committed bool) (value []byte, err error) {
	if committed {
		return appState.getCommittedVersioned(key, height)
	} else {
		return appState.getVersioned(key, height)
	}
}

func (appState *AppState) getVersioned(key []byte, height int64) (value []byte, err error) {
	versionsKeyStr := string(key) + "|versions"
	versionsKey := []byte(versionsKeyStr)

	var versions []int64
	var existInUncommittedState bool
	versions, existInUncommittedState = appState.uncommittedVersionsState[versionsKeyStr]
	if !existInUncommittedState {
		keyVersionsProtobuf, err := appState.db.Get(versionsKey)
		if err != nil {
			return nil, err
		}
		if keyVersionsProtobuf != nil {
			var keyVersions data.KeyVersions
			err = proto.Unmarshal([]byte(keyVersionsProtobuf), &keyVersions)
			if err != nil {
				return nil, err
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
		var exist bool
		value, exist = appState.uncommittedState[keyWithVersionStr]
		if !exist {
			keyWithVersion := []byte(keyWithVersionStr)
			value, err = appState.db.Get(keyWithVersion)
			if err != nil {
				return nil, err
			}
		}
	} else {
		keyWithVersion := []byte(keyWithVersionStr)
		value, err = appState.db.Get(keyWithVersion)
		if err != nil {
			return nil, err
		}
	}

	return value, nil
}

func (appState *AppState) getCommittedVersioned(key []byte, height int64) (value []byte, err error) {
	versionsKeyStr := string(key) + "|versions"
	versionsKey := []byte(versionsKeyStr)

	var versions []int64
	keyVersionsProtobuf, err := appState.db.Get(versionsKey)
	if err != nil {
		return nil, err
	}
	var keyVersions data.KeyVersions
	err = proto.Unmarshal([]byte(keyVersionsProtobuf), &keyVersions)
	if err != nil {
		return nil, err
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

	value, err = appState.db.Get(keyWithVersion)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (appState *AppState) Has(key []byte, committed bool) (bool, error) {
	if committed {
		return appState.hasCommitted(key)
	} else {
		return appState.has(key)
	}
}

func (appState *AppState) has(key []byte) (bool, error) {
	value, existInUncommittedState := appState.uncommittedState[string(key)]
	if existInUncommittedState {
		if value != nil {
			return true, nil
		} else {
			return false, nil
		}
	}
	return appState.db.Has(key)
}

func (appState *AppState) hasCommitted(key []byte) (bool, error) {
	return appState.db.Has(key)
}

func (appState *AppState) HasVersioned(key []byte, committed bool) (bool, error) {
	if committed {
		return appState.hasCommittedVersioned(key)
	} else {
		return appState.hasVersioned(key)
	}
}

func (appState *AppState) hasVersioned(key []byte) (bool, error) {
	versionsKeyStr := string(key) + "|versions"
	versionsKey := []byte(versionsKeyStr)

	value, existInUncommittedState := appState.uncommittedVersionsState[versionsKeyStr]
	if existInUncommittedState {
		if value != nil {
			return true, nil
		} else {
			return false, nil
		}
	}

	return appState.db.Has(versionsKey)
}

func (appState *AppState) hasCommittedVersioned(key []byte) (bool, error) {
	versionsKeyStr := string(key) + "|versions"
	versionsKey := []byte(versionsKeyStr)
	return appState.db.Has(versionsKey)
}

func (appState *AppState) Delete(key []byte) error {
	hasKey, err := appState.has(key)
	if err != nil {
		return err
	}
	if !hasKey {
		return nil
	}

	appState.HasHashData = true
	appState.HashDigest.Write(key)
	appState.HashDigest.Write(actionDelete)

	appState.uncommittedState[string(key)] = nil

	return nil
}

func (appState *AppState) DeleteVersioned(key []byte) error {
	hasVersionedKey, err := appState.hasVersioned(key)
	if err != nil {
		return err
	}
	if !hasVersionedKey {
		return nil
	}
	appState.SetVersioned(key, nil)

	return nil
}

func (appState *AppState) Save() error {
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
			return err
		}
		batch.Set([]byte(key), value)
	}

	batch.WriteSync()

	appState.uncommittedState = make(map[string][]byte)
	appState.uncommittedVersionsState = make(map[string][]int64)

	return nil
}
