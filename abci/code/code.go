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
	OK                                                            uint32 = 0
	EncodingError                                                 uint32 = 1
	DecodingError                                                 uint32 = 2
	BadNonce                                                      uint32 = 3
	Unauthorized                                                  uint32 = 4
	UnmarshalError                                                uint32 = 5
	MarshalError                                                  uint32 = 6
	RequestIDNotFound                                             uint32 = 7
	RequestIsClosed                                               uint32 = 8
	RequestIsTimedOut                                             uint32 = 9
	RequestIsCompleted                                            uint32 = 10
	DuplicateServiceID                                            uint32 = 11
	TokenAccountNotFound                                          uint32 = 12
	TokenNotEnough                                                uint32 = 13
	InvalidTransactionFormat                                      uint32 = 14
	MethodCanNotBeEmpty                                           uint32 = 15
	DuplicateIdPResponse                                          uint32 = 16
	AALError                                                      uint32 = 17
	IALError                                                      uint32 = 18
	DuplicateNodeID                                               uint32 = 19
	WrongRole                                                     uint32 = 20
	DuplicateNamespace                                            uint32 = 21
	NamespaceNotFound                                             uint32 = 22
	DuplicateRequestID                                            uint32 = 23
	NodeIDNotFound                                                uint32 = 24
	DuplicatePublicKey                                            uint32 = 25
	DuplicateAccessorID                                           uint32 = 26
	DuplicateAccessorGroupID                                      uint32 = 27
	AccessorGroupIDNotFound                                       uint32 = 28
	RequestIsNotCompleted                                         uint32 = 29
	RequestIsNotSpecial                                           uint32 = 30
	InvalidMinIdp                                                 uint32 = 31
	NodeIDDoesNotExistInASList                                    uint32 = 32
	AsIDDoesNotExistInASList                                      uint32 = 33
	ServiceIDNotFound                                             uint32 = 34
	InvalidMode                                                   uint32 = 35
	HashIDNotFound                                                uint32 = 36
	DuplicateASInDataRequest                                      uint32 = 39
	DuplicateASResponse                                           uint32 = 40
	DuplicateServiceIDInDataRequest                               uint32 = 41
	ServiceDestinationNotFound                                    uint32 = 42
	DataRequestIsCompleted                                        uint32 = 43
	NotFirstIdP                                                   uint32 = 44
	AccessorIDNotFound                                            uint32 = 45
	NotOwnerOfAccessor                                            uint32 = 46
	NoPermissionForRegisterServiceDestination                     uint32 = 47
	IncompleteValidList                                           uint32 = 48
	UnknownMethod                                                 uint32 = 49
	InvalidKeyFormat                                              uint32 = 50
	UnsupportedKeyType                                            uint32 = 51
	UnknownKeyType                                                uint32 = 52
	NDIDisAlreadyExisted                                          uint32 = 53
	NoPermissionForSetMqAddresses                                 uint32 = 54
	NoPermissionForCallNDIDMethod                                 uint32 = 55
	NoPermissionForCallIdPMethod                                  uint32 = 56
	NoPermissionForCallASMethod                                   uint32 = 57
	NoPermissionForCallRPandIdPMethod                             uint32 = 58
	VerifySignatureError                                          uint32 = 59
	NotOwnerOfRequest                                             uint32 = 60
	CannotGetPublicKeyFromParam                                   uint32 = 61
	CannotGetMasterPublicKeyFromNodeID                            uint32 = 62
	CannotGetPublicKeyFromNodeID                                  uint32 = 63
	TimeOutBlockIsMustGreaterThanZero                             uint32 = 64
	RSAKeyLengthTooShort                                          uint32 = 65
	RegisterIdentityIsTimedOut                                    uint32 = 66
	AmountMustBeGreaterOrEqualToZero                              uint32 = 67
	NodeIsNotActive                                               uint32 = 68
	ServiceIsNotActive                                            uint32 = 69
	ServiceDestinationIsNotActive                                 uint32 = 70
	ServiceDestinationIsNotApprovedByNDID                         uint32 = 71
	NodeIDIsAlreadyAssociatedWithProxyNode                        uint32 = 72
	NodeIDisProxyNode                                             uint32 = 73
	NodeIDHasNotBeenAssociatedWithProxyNode                       uint32 = 74
	ProxyNodeNotFound                                             uint32 = 75
	NodeIDDoesNotExistInIdPList                                   uint32 = 76
	ProxyNodeIsNotActive                                          uint32 = 77
	NodeIDInIdPListIsNotActive                                    uint32 = 78
	NodeIDInASListIsNotActive                                     uint32 = 79
	RoleIsNotAS                                                   uint32 = 80
	RequestIsNotClosed                                            uint32 = 81
	ChainIsDisabled                                               uint32 = 82
	ChainIsNotInitialized                                         uint32 = 83
	DuplicateNonce                                                uint32 = 84
	NotExistValidator                                             uint32 = 85
	IdentityAlreadyExisted                                        uint32 = 86
	IdentityCannotBeEmpty                                         uint32 = 87
	GotRefGroupCodeAndIdentity                                    uint32 = 88
	RefGroupNotFound                                              uint32 = 89
	IdentityNotFoundInThisIdP                                     uint32 = 90
	InvalidPurpose                                                uint32 = 91
	RequestIsAlreadyUsed                                          uint32 = 92
	RefGroupCodeCannotBeEmpty                                     uint32 = 93
	AllAccessorMustHaveSameRefGroupCode                           uint32 = 94
	AccessorNotFoundInThisIdP                                     uint32 = 95
	AccessorIDCannotBeEmpty                                       uint32 = 97
	AccessorPublicKeyCannotBeEmpty                                uint32 = 98
	AccessorTypeCannotBeEmpty                                     uint32 = 99
	InvalidNamespace                                              uint32 = 100
	IdentifierCountIsGreaterThanAllowedIdentifierCount            uint32 = 101
	IalMustBeGreaterOrEqualMinIal                                 uint32 = 102
	CannotRevokeAllAccessorsInThisIdP                             uint32 = 103
	DuplicateIdentifier                                           uint32 = 104
	NewModeListMustBeHigherThanCurrentModeList                    uint32 = 105
	InvalidErrorCode                                              uint32 = 106
	NodeNotInWhitelist                                            uint32 = 107
	AppStateError                                                 uint32 = 108
	RequestCannotBeFulfilled                                      uint32 = 109
	DataRequestCannotBeFulfilled                                  uint32 = 110
	IdPCreateRequestWithDataRequestNotAllowed                     uint32 = 111
	IdPCreateRequestMode1And2NotAllowed                           uint32 = 112
	ServicePriceCeilingListNotFound                               uint32 = 113
	ServicePriceCeilingNotFound                                   uint32 = 114
	ServiceMaxPriceGreaterThanPriceCeiling                        uint32 = 115
	ServicePriceEffectiveDatetimeBeforeAllowed                    uint32 = 116
	ServicePriceMinCannotBeGreaterThanMax                         uint32 = 117
	NothingToUpdateIdentity                                       uint32 = 118
	NoPermissionForCallRPMethod                                   uint32 = 119
	DuplicateMessageID                                            uint32 = 120
	RequestTypeAlreadyExists                                      uint32 = 121
	RequestTypeDoesNotExist                                       uint32 = 122
	SuppressedIdentityModificationNotificationNodeIDAlreadyExists uint32 = 123
	SuppressedIdentityModificationNotificationNodeIDDoesNotExist  uint32 = 124
	NotIdPNode                                                    uint32 = 125
	ChainIsAlreadyInitialized                                     uint32 = 126
	ChainIdMismatch                                               uint32 = 127

	UnknownError uint32 = 999
)
