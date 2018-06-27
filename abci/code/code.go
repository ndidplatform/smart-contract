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

package code

// Return codes for return result
const (
	OK                              uint32 = 0
	EncodingError                   uint32 = 1
	DecodingError                   uint32 = 2
	BadNonce                        uint32 = 3
	Unauthorized                    uint32 = 4
	UnmarshalError                  uint32 = 5
	MarshalError                    uint32 = 6
	RequestIDNotFound               uint32 = 7
	RequestIsClosed                 uint32 = 8
	RequestIsTimedOut               uint32 = 9
	RequestIsCompleted              uint32 = 10
	DuplicateServiceID              uint32 = 11
	TokenAccountNotFound            uint32 = 12
	TokenNotEnough                  uint32 = 13
	WrongTransactionFormat          uint32 = 14
	MethodCanNotBeEmpty             uint32 = 15
	DuplicateResponse               uint32 = 16
	AALError                        uint32 = 17
	IALError                        uint32 = 18
	DuplicateNodeID                 uint32 = 19
	WrongRole                       uint32 = 20
	DuplicateNamespace              uint32 = 21
	NamespaceNotFound               uint32 = 22
	DuplicateRequestID              uint32 = 23
	NodeIDNotFound                  uint32 = 24
	DuplicatePublicKey              uint32 = 25
	DuplicateAccessorID             uint32 = 26
	DuplicateAccessorGroupID        uint32 = 27
	AccessorGroupIDNotFound         uint32 = 28
	RequestIsNotCompleted           uint32 = 29
	RequestIsNotSpecial             uint32 = 30
	InvalidMinIdp                   uint32 = 31
	NodeIDIsNotExistInASList        uint32 = 32
	AsIDIsNotExistInASList          uint32 = 33
	ServiceIDNotFound               uint32 = 34
	InvalidMode                     uint32 = 35
	HashIDNotFound                  uint32 = 36
	DuplicateIdentityProof          uint32 = 37
	WrongIdentityProof              uint32 = 38
	DuplicateASInDataRequest        uint32 = 39
	DuplicateAnsweredAsIDList       uint32 = 40
	DuplicateServiceIDInDataRequest uint32 = 41
	ServiceDestinationNotFound      uint32 = 42
	DataRequestIsCompleted          uint32 = 43
	NotFirstIdP                     uint32 = 44
	AccessorIDNotFound              uint32 = 45
	NotOwnerOfAccessor              uint32 = 46
)
