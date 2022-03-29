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
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	dbm "github.com/tendermint/tm-db"
)

var testDb dbm.DB

func TestMain(m *testing.M) {
	var err error

	var dbType = "goleveldb"
	dbDir, err := os.MkdirTemp(os.TempDir(), "ndid-smart-contract-unit-test-")
	if err != nil {
		panic(fmt.Errorf("Could not create temp DB directory: %v", err.Error()))
	}

	// setup
	name := "didDB"
	testDb, err = dbm.NewDB(name, dbm.BackendType(dbType), dbDir)
	if err != nil {
		panic(fmt.Errorf("Could not create DB instance: %v", err.Error()))
	}

	code := m.Run()

	// teardown
	err = os.RemoveAll(dbDir)
	if err != nil {
		panic(fmt.Errorf("Could not delete temp DB directory: %v", err.Error()))
	}

	os.Exit(code)
}

func TestSet1(t *testing.T) {
	var err error
	var value []byte

	key := []byte("testsetkey1")
	valueToSet := []byte("value1")

	appState, err := NewAppState(testDb)
	if err != nil {
		t.Fatalf("error new app state: %+v", err)
	}

	value, err = appState.Get(key, true)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Nil(t, value)

	value, err = appState.Get(key, false)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Nil(t, value)

	appState.Set(key, valueToSet)

	value, err = appState.Get(key, true)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Nil(t, value)

	value, err = appState.Get(key, false)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Equal(t, valueToSet, value)
}

func TestSetAndSave1(t *testing.T) {
	var err error
	var value []byte

	key := []byte("testsetandsavekey1")
	valueToSet := []byte("value1")

	appState, err := NewAppState(testDb)
	if err != nil {
		t.Fatalf("error new app state: %+v", err)
	}

	value, err = appState.Get(key, true)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Nil(t, value)

	value, err = appState.Get(key, false)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Nil(t, value)

	appState.Set(key, valueToSet)

	value, err = appState.Get(key, true)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Nil(t, value)

	value, err = appState.Get(key, false)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Equal(t, valueToSet, value)

	err = appState.Save()
	if err != nil {
		t.Fatalf("error save: %+v", err)
	}

	value, err = appState.Get(key, true)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Equal(t, valueToSet, value)
}

func TestSetVersioned1(t *testing.T) {
	var err error
	var value []byte

	key := []byte("testsetversionedkey1")
	valueToSet := []byte("value1")

	appState, err := NewAppState(testDb)
	if err != nil {
		t.Fatalf("error new app state: %+v", err)
	}

	appState.CurrentBlockHeight = 1

	value, err = appState.GetVersioned(key, 1, true)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Nil(t, value)

	value, err = appState.GetVersioned(key, 1, false)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Nil(t, value)

	appState.SetVersioned(key, valueToSet)

	value, err = appState.GetVersioned(key, 1, true)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Nil(t, value)

	value, err = appState.GetVersioned(key, 1, false)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Equal(t, valueToSet, value)

	value, err = appState.GetVersioned(key, 0, false)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Equal(t, valueToSet, value)
}

func TestSetVersioned2(t *testing.T) {
	var err error
	var value []byte

	key := []byte("testsetversionedkey2")
	valueToSetVersions1 := []byte("value1")
	valueToSetVersions2 := []byte("value2")

	appState, err := NewAppState(testDb)
	if err != nil {
		t.Fatalf("error new app state: %+v", err)
	}

	appState.CurrentBlockHeight = 1

	value, err = appState.GetVersioned(key, 1, true)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Nil(t, value)

	value, err = appState.GetVersioned(key, 1, false)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Nil(t, value)

	appState.SetVersioned(key, valueToSetVersions1)

	value, err = appState.GetVersioned(key, 1, true)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Nil(t, value)

	value, err = appState.GetVersioned(key, 1, false)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Equal(t, valueToSetVersions1, value)

	appState.CurrentBlockHeight = 2

	appState.SetVersioned(key, valueToSetVersions2)

	value, err = appState.GetVersioned(key, 1, false)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Equal(t, valueToSetVersions1, value)

	value, err = appState.GetVersioned(key, 2, false)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Equal(t, valueToSetVersions2, value)

	value, err = appState.GetVersioned(key, 0, false)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Equal(t, valueToSetVersions2, value)
}

func TestSetVersionedAndSave1(t *testing.T) {
	var err error
	var value []byte

	key := []byte("testsetversionedandsavekey1")
	valueToSetVersions1 := []byte("value1")
	valueToSetVersions2 := []byte("value2")

	appState, err := NewAppState(testDb)
	if err != nil {
		t.Fatalf("error new app state: %+v", err)
	}

	appState.CurrentBlockHeight = 1

	value, err = appState.GetVersioned(key, 1, true)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Nil(t, value)

	value, err = appState.GetVersioned(key, 1, false)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Nil(t, value)

	appState.SetVersioned(key, valueToSetVersions1)

	value, err = appState.GetVersioned(key, 1, true)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Nil(t, value)

	value, err = appState.GetVersioned(key, 1, false)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Equal(t, valueToSetVersions1, value)

	err = appState.Save()
	if err != nil {
		t.Fatalf("error save: %+v", err)
	}

	appState.CurrentBlockHeight = 2

	appState.SetVersioned(key, valueToSetVersions2)

	value, err = appState.GetVersioned(key, 1, true)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Equal(t, valueToSetVersions1, value)

	value, err = appState.GetVersioned(key, 0, true)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Equal(t, valueToSetVersions1, value)

	value, err = appState.GetVersioned(key, 1, false)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Equal(t, valueToSetVersions1, value)

	value, err = appState.GetVersioned(key, 2, true)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Equal(t, valueToSetVersions1, value)

	value, err = appState.GetVersioned(key, 2, false)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Equal(t, valueToSetVersions2, value)

	value, err = appState.GetVersioned(key, 0, false)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Equal(t, valueToSetVersions2, value)
}

func TestHas1(t *testing.T) {
	var err error
	var exist bool

	key := []byte("testhaskey1")
	valueToSet := []byte("value1")

	appState, err := NewAppState(testDb)
	if err != nil {
		t.Fatalf("error new app state: %+v", err)
	}

	exist, err = appState.Has(key, true)
	if err != nil {
		t.Fatalf("error has: %+v", err)
	}
	assert.False(t, exist)

	exist, err = appState.Has(key, false)
	if err != nil {
		t.Fatalf("error has: %+v", err)
	}
	assert.False(t, exist)

	appState.Set(key, valueToSet)

	exist, err = appState.Has(key, true)
	if err != nil {
		t.Fatalf("error has: %+v", err)
	}
	assert.False(t, exist)

	exist, err = appState.Has(key, false)
	if err != nil {
		t.Fatalf("error has: %+v", err)
	}
	assert.True(t, exist)
}

func TestHas2(t *testing.T) {
	var err error
	var exist bool

	key := []byte("testhaskey2")
	valueToSet := []byte("value1")

	appState, err := NewAppState(testDb)
	if err != nil {
		t.Fatalf("error new app state: %+v", err)
	}

	exist, err = appState.Has(key, true)
	if err != nil {
		t.Fatalf("error has: %+v", err)
	}
	assert.False(t, exist)

	exist, err = appState.Has(key, false)
	if err != nil {
		t.Fatalf("error has: %+v", err)
	}
	assert.False(t, exist)

	appState.Set(key, valueToSet)

	appState.Delete(key)

	exist, err = appState.Has(key, true)
	if err != nil {
		t.Fatalf("error has: %+v", err)
	}
	assert.False(t, exist)

	exist, err = appState.Has(key, false)
	if err != nil {
		t.Fatalf("error has: %+v", err)
	}
	assert.False(t, exist)
}

func TestHasAndSave1(t *testing.T) {
	var err error
	var exist bool

	key := []byte("testhasandsavekey1")
	valueToSet := []byte("value1")

	appState, err := NewAppState(testDb)
	if err != nil {
		t.Fatalf("error new app state: %+v", err)
	}

	exist, err = appState.Has(key, true)
	if err != nil {
		t.Fatalf("error has: %+v", err)
	}
	assert.False(t, exist)

	exist, err = appState.Has(key, false)
	if err != nil {
		t.Fatalf("error has: %+v", err)
	}
	assert.False(t, exist)

	appState.Set(key, valueToSet)

	appState.Save()

	exist, err = appState.Has(key, true)
	if err != nil {
		t.Fatalf("error has: %+v", err)
	}
	assert.True(t, exist)

	exist, err = appState.Has(key, false)
	if err != nil {
		t.Fatalf("error has: %+v", err)
	}
	assert.True(t, exist)
}

func TestHasVersioned1(t *testing.T) {
	var err error
	var exist bool

	key := []byte("testhasversionedkey1")
	valueToSet := []byte("value1")

	appState, err := NewAppState(testDb)
	if err != nil {
		t.Fatalf("error new app state: %+v", err)
	}

	appState.CurrentBlockHeight = 1

	exist, err = appState.HasVersioned(key, true)
	if err != nil {
		t.Fatalf("error has: %+v", err)
	}
	assert.False(t, exist)

	exist, err = appState.HasVersioned(key, false)
	if err != nil {
		t.Fatalf("error has: %+v", err)
	}
	assert.False(t, exist)

	appState.SetVersioned(key, valueToSet)

	exist, err = appState.HasVersioned(key, true)
	if err != nil {
		t.Fatalf("error has: %+v", err)
	}
	assert.False(t, exist)

	exist, err = appState.HasVersioned(key, false)
	if err != nil {
		t.Fatalf("error has: %+v", err)
	}
	assert.True(t, exist)
}

func TestDelete1(t *testing.T) {
	var err error
	var value []byte

	key := []byte("testdeletekey1")
	valueToSet := []byte("value1")

	appState, err := NewAppState(testDb)
	if err != nil {
		t.Fatalf("error new app state: %+v", err)
	}

	value, err = appState.Get(key, true)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Nil(t, value)

	value, err = appState.Get(key, false)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Nil(t, value)

	appState.Set(key, valueToSet)

	value, err = appState.Get(key, true)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Nil(t, value)

	value, err = appState.Get(key, false)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Equal(t, valueToSet, value)

	// delete
	err = appState.Delete(key)
	if err != nil {
		t.Fatalf("error delete: %+v", err)
	}

	value, err = appState.Get(key, false)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Nil(t, value)
}

func TestDeleteAndSave1(t *testing.T) {
	var err error
	var value []byte

	key := []byte("testdeleteandsavekey1")
	valueToSet := []byte("value1")

	appState, err := NewAppState(testDb)
	if err != nil {
		t.Fatalf("error new app state: %+v", err)
	}

	value, err = appState.Get(key, true)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Nil(t, value)

	value, err = appState.Get(key, false)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Nil(t, value)

	appState.Set(key, valueToSet)

	value, err = appState.Get(key, true)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Nil(t, value)

	value, err = appState.Get(key, false)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Equal(t, valueToSet, value)

	err = appState.Save()
	if err != nil {
		t.Fatalf("error save: %+v", err)
	}

	// delete
	err = appState.Delete(key)
	if err != nil {
		t.Fatalf("error delete: %+v", err)
	}

	value, err = appState.Get(key, true)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Equal(t, valueToSet, value)

	value, err = appState.Get(key, false)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Nil(t, value)

	err = appState.Save()
	if err != nil {
		t.Fatalf("error save: %+v", err)
	}

	value, err = appState.Get(key, true)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Nil(t, value)
}

func TestDeleteVersioned1(t *testing.T) {
	var err error
	var value []byte

	key := []byte("testdeleteversionedkey1")
	valueToSet := []byte("value1")

	appState, err := NewAppState(testDb)
	if err != nil {
		t.Fatalf("error new app state: %+v", err)
	}

	appState.CurrentBlockHeight = 1

	value, err = appState.GetVersioned(key, 1, true)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Nil(t, value)

	value, err = appState.GetVersioned(key, 1, false)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Nil(t, value)

	appState.SetVersioned(key, valueToSet)

	value, err = appState.GetVersioned(key, 1, true)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Nil(t, value)

	value, err = appState.GetVersioned(key, 1, false)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Equal(t, valueToSet, value)

	value, err = appState.GetVersioned(key, 0, false)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Equal(t, valueToSet, value)

	err = appState.DeleteVersioned(key)
	if err != nil {
		t.Fatalf("error delete: %+v", err)
	}

	value, err = appState.GetVersioned(key, 1, false)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Nil(t, value)
}

func TestDeleteVersionedAndSave1(t *testing.T) {
	var err error
	var value []byte

	key := []byte("testdeleteversionedandsavekey1")
	valueToSetVersions1 := []byte("value1")
	valueToSetVersions2 := []byte("value2")

	appState, err := NewAppState(testDb)
	if err != nil {
		t.Fatalf("error new app state: %+v", err)
	}

	appState.CurrentBlockHeight = 1

	value, err = appState.GetVersioned(key, 1, true)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Nil(t, value)

	value, err = appState.GetVersioned(key, 1, false)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Nil(t, value)

	appState.SetVersioned(key, valueToSetVersions1)

	value, err = appState.GetVersioned(key, 1, true)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Nil(t, value)

	value, err = appState.GetVersioned(key, 1, false)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Equal(t, valueToSetVersions1, value)

	value, err = appState.GetVersioned(key, 0, false)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Equal(t, valueToSetVersions1, value)

	err = appState.DeleteVersioned(key)
	if err != nil {
		t.Fatalf("error delete: %+v", err)
	}

	value, err = appState.GetVersioned(key, 1, false)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Nil(t, value)

	appState.SetVersioned(key, valueToSetVersions1)

	err = appState.Save()
	if err != nil {
		t.Fatalf("error save: %+v", err)
	}

	appState.CurrentBlockHeight = 2

	appState.SetVersioned(key, valueToSetVersions2)

	err = appState.DeleteVersioned(key)
	if err != nil {
		t.Fatalf("error delete: %+v", err)
	}

	value, err = appState.GetVersioned(key, 1, false)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Equal(t, valueToSetVersions1, value)

	value, err = appState.GetVersioned(key, 2, false)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Nil(t, value)

	value, err = appState.GetVersioned(key, 0, false)
	if err != nil {
		t.Fatalf("error get: %+v", err)
	}
	assert.Nil(t, value)
}
