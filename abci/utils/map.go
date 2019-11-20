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

package utils

import (
	"sync"
)

type StringMap struct {
	sync.RWMutex
	internal map[string]string
}

func NewStringMap() *StringMap {
	return &StringMap{
		internal: make(map[string]string),
	}
}

func (rm *StringMap) Load(key string) (value string, ok bool) {
	rm.RLock()
	result, ok := rm.internal[key]
	rm.RUnlock()
	return result, ok
}

func (rm *StringMap) Delete(key string) {
	rm.Lock()
	delete(rm.internal, key)
	rm.Unlock()
}

func (rm *StringMap) Store(key, value string) {
	rm.Lock()
	rm.internal[key] = value
	rm.Unlock()
}

type StringByteArrayMap struct {
	sync.RWMutex
	internal map[string][]byte
}

func NewStringByteArrayMap() *StringByteArrayMap {
	return &StringByteArrayMap{
		internal: make(map[string][]byte),
	}
}

func (rm *StringByteArrayMap) Load(key string) (value []byte, ok bool) {
	rm.RLock()
	result, ok := rm.internal[key]
	rm.RUnlock()
	return result, ok
}

func (rm *StringByteArrayMap) Delete(key string) {
	rm.Lock()
	delete(rm.internal, key)
	rm.Unlock()
}

func (rm *StringByteArrayMap) Store(key string, value []byte) {
	rm.Lock()
	rm.internal[key] = value
	rm.Unlock()
}
