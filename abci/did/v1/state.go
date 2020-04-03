package did

import (
	"encoding/binary"
	"encoding/json"
	"strconv"

	"github.com/golang/protobuf/proto"
	dbm "github.com/tendermint/tm-db"

	"github.com/ndidplatform/smart-contract/v4/abci/utils"
	"github.com/ndidplatform/smart-contract/v4/protos/data"
)

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
	CurrentBlockHeight       int64
	HashData                 []byte
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
		HashData:                 make([]byte, 0),
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

func (appState *AppState) Set(key, value []byte) {
	appState.HashData = append(appState.HashData, key...)
	appState.HashData = append(appState.HashData, value...)

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
		appState.HashData = append(appState.HashData, versionsKey...)
		versionBytes := make([]byte, 8)
		for _, version := range versions {
			binary.BigEndian.PutUint64(versionBytes, uint64(version))
			appState.HashData = append(appState.HashData, versionBytes...)
		}

		appState.uncommittedVersionsState[versionsKeyStr] = append(versions, appState.CurrentBlockHeight)
	}

	keyWithVersionStr := string(key) + "|" + strconv.FormatInt(appState.CurrentBlockHeight, 10)

	appState.HashData = append(appState.HashData, key...)
	appState.HashData = append(appState.HashData, value...)

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
		value = appState.uncommittedState[keyWithVersionStr]
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
	_, existInUncommittedState := appState.uncommittedState[string(key)]
	if existInUncommittedState {
		return true, nil
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

	_, existInUncommittedState := appState.uncommittedVersionsState[versionsKeyStr]
	if existInUncommittedState {
		return true, nil
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
	appState.HashData = append(appState.HashData, key...)
	appState.HashData = append(appState.HashData, []byte("delete")...) // Remove or replace with something else?

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
