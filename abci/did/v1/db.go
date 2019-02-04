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
	"strconv"

	"github.com/golang/protobuf/proto"
	"github.com/ndidplatform/smart-contract/abci/utils"
	"github.com/ndidplatform/smart-contract/protos/data"
)

func (app *DIDApplication) SetStateDB(key, value []byte) {
	app.HashData = append(app.HashData, key...)
	app.HashData = append(app.HashData, value...)

	app.UncommittedState[string(key)] = value
}

func (app *DIDApplication) SetVersionedStateDB(key, value []byte) {
	versionsKeyStr := string(key) + "|versions"
	versionsKey := []byte(versionsKeyStr)

	var versions []int64
	var existInUncommittedState bool
	versions, existInUncommittedState = app.UncommittedVersionsState[versionsKeyStr]
	if !existInUncommittedState {
		keyVersionsProtobuf := app.state.db.Get(versionsKey)
		if keyVersionsProtobuf != nil {
			var keyVersions data.KeyVersions
			err := proto.Unmarshal([]byte(keyVersionsProtobuf), &keyVersions)
			if err != nil {
				panic(err) // Should panic or return err?
			}
			versions = keyVersions.Versions
		}
	}

	if len(versions) == 0 || versions[len(versions)-1] != app.CurrentBlock {
		app.HashData = append(app.HashData, versionsKey...)
		versionBytes := make([]byte, 8)
		for _, version := range versions {
			binary.BigEndian.PutUint64(versionBytes, uint64(version))
			app.HashData = append(app.HashData, versionBytes...)
		}

		app.UncommittedVersionsState[versionsKeyStr] = append(versions, app.CurrentBlock)
	}

	keyWithVersionStr := string(key) + "|" + strconv.FormatInt(app.CurrentBlock, 10)

	app.HashData = append(app.HashData, key...)
	app.HashData = append(app.HashData, value...)

	app.UncommittedState[keyWithVersionStr] = value
}

func (app *DIDApplication) GetStateDB(key []byte) (err error, value []byte) {
	var existInUncommittedState bool
	value, existInUncommittedState = app.UncommittedState[string(key)]
	if !existInUncommittedState {
		value = app.state.db.Get(key)
	}

	return nil, value
}

func (app *DIDApplication) GetVersionedStateDB(key []byte, height int64) (err error, value []byte) {
	versionsKeyStr := string(key) + "|versions"
	versionsKey := []byte(versionsKeyStr)

	var versions []int64
	var existInUncommittedState bool
	versions, existInUncommittedState = app.UncommittedVersionsState[versionsKeyStr]
	if !existInUncommittedState {
		keyVersionsProtobuf := app.state.db.Get(versionsKey)
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
		value = app.UncommittedState[keyWithVersionStr]
	} else {
		keyWithVersion := []byte(keyWithVersionStr)
		value = app.state.db.Get(keyWithVersion)
	}

	return nil, value
}

func (app *DIDApplication) GetCommittedStateDB(key []byte) (err error, value []byte) {
	value = app.state.db.Get(key)
	return nil, value
}

func (app *DIDApplication) GetCommittedVersionedStateDB(key []byte, height int64) (err error, value []byte) {
	versionsKeyStr := string(key) + "|versions"
	versionsKey := []byte(versionsKeyStr)

	var versions []int64
	keyVersionsProtobuf := app.state.db.Get(versionsKey)
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

	value = app.state.db.Get(keyWithVersion)
	return nil, value
}

func (app *DIDApplication) HasStateDB(key []byte) bool {
	_, existInUncommittedState := app.UncommittedState[string(key)]
	if existInUncommittedState {
		return true
	}
	return app.state.db.Has(key)
}

func (app *DIDApplication) HasVersionedStateDB(key []byte) bool {
	versionsKeyStr := string(key) + "|versions"
	versionsKey := []byte(versionsKeyStr)

	_, existInUncommittedState := app.UncommittedVersionsState[versionsKeyStr]
	if existInUncommittedState {
		return true
	}

	return app.state.db.Has(versionsKey)
}

func (app *DIDApplication) DeleteStateDB(key []byte) {
	app.HashData = append(app.HashData, key...)
	app.HashData = append(app.HashData, []byte("delete")...) // Remove or replace with something else?

	app.UncommittedState[string(key)] = nil
}

func (app *DIDApplication) DeleteVersionedStateDB(key []byte) {
	if !app.HasVersionedStateDB(key) {
		return
	}
	app.SetVersionedStateDB(key, nil)
}

func (app *DIDApplication) SaveDBState() {
	batch := app.state.db.NewBatch()

	for key := range app.UncommittedState {
		value := app.UncommittedState[key]
		if value != nil {
			batch.Set([]byte(key), value)
		} else {
			batch.Delete([]byte(key))
		}
	}

	for key := range app.UncommittedVersionsState {
		versions := app.UncommittedVersionsState[key]
		var keyVersions data.KeyVersions
		keyVersions.Versions = versions
		value, err := utils.ProtoDeterministicMarshal(&keyVersions)
		if err != nil {
			panic(err) // Should panic or return err?
		}
		batch.Set([]byte(key), value)
	}

	batch.WriteSync()

	app.UncommittedState = make(map[string][]byte)
	app.UncommittedVersionsState = make(map[string][]int64)
}
