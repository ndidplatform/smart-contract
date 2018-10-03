# Changelog

## 0.9.0 (October 4, 2018)

BREAKING CHANGES:

- [DeliverTx] Delete `node_id` property from parameters of `SetMqAddresses`
- [DeliverTx] Add `idp_id_list` property to parameters of `CreateRequest`
- [DeliverTx] Add `purpose` property to parameters of `CreateRequest`
- [Query] Return `purpose` instead `special` in result of `GetRequestDetail`.
- [CheckTx] Check proxy node is active when creating a transaction.
- [DeliverTx] Add `data_schema` and `data_schema_version` property to parameters of `AddService` and `UpdateService`.
- [Query] Add `data_schema` and `data_schema_version` property to result of `GetServiceDetail`.

IMPROVEMENTS:

- Refactor code.
- [Query] Add `idp_id_list` property to result of `GetRequestDetail`.
- [Key/Value store] Add accessorID in accessorGroupID relation.
- [Query] Add new function (`GetAccessorsInAccessorGroup`).
- [Query] Add `creation_block_height` property to result of `GetRequestDetail`.

BUG FIXES:

- [DeliverTx] Fix unmarshal error in `UpdateNodeProxyNode`.
- [DeliverTx] Remove invalid key of `MqAddresses` in `AddNodeToProxyNode`.

## 0.8.0 (September 23, 2018)

BREAKING CHANGES:

- Rename functions
  - `RegisterMsqAddress` to `SetMqAddresses`
  - `RegisterMsqDestination` to `RegisterIdentity`
  - `ClearRegisterMsqDestinationTimeout` to `ClearRegisterIdentityTimeout`
  - `SetTimeOutBlockRegisterMsqDestination` to `SetTimeOutBlockRegisterIdentity`
  - `CreateIdentity` to `RegisterAccessor`
  - `GetMsqAddress` to `GetMqAddresses`
- [DeliverTx] Update `RegisterNode` to be able to register proxy node.
- [DeliverTx] Add new functions (`AddNodeToProxyNode`, `UpdateNodeProxyNode` and `RemoveNodeFromProxyNode`).
- [Query] Add new functions (`GetNodesBehindProxyNode`, `GetNodeIDList`).
- [Key/Value store] Change all stored data format in app state DB from `JSON` to `Protobuf`.

IMPROVEMENTS:

- Change Tx and Query input format from base64 string to byte array.
- [Query] Add `GetIdpNodesInfo` function for get IdP node with mq addresses.
- [Query] Add `GetAsNodesInfoByServiceId` function for get AS node with mq addresses.
- [Query] Add mq addresses property to result of `GetNodeInfo`.
- [Query] Add `requester_node_id` to request returns from calling `GetRequestDetail`
 
BUG FIXES:

- [Query] Node name is invalid after update node (GetAsNodesByServiceId)
- [DeliverTx] Return invalid code and log when call `UpdateNode` by NDID

## 0.7.2 (August 22, 2018)

IMPROVEMENTS:

- [Query] Return log "not found" and value in JSON format when query not found
- [DeliverTx] Check service is active (RegisterServiceDestination)

## 0.7.1 (August 16, 2018)

IMPROVEMENTS:

- [Docker] Set umask to 027 and use user nobody to run service
- [Docker] Add security_opt: no-new-priviledges in docker-compose file

## 0.7.0 (August 15, 2018)

BREAKING CHANGES:

- [DeliverTx] Remove DisableMsqDestination, DisableAccessorMethod, EnableMsqDestination and EnableAccessorMethod functions

IMPROVEMENTS:

- [CheckTx] Check amount value to be greater than or equal to zero (token function)
- [CheckTx] Check node is active when creating a transaction
- [DeliverTx] Check service is active (SignData)
- [DeliverTx] Check service destination is active (SignData)
- [DeliverTx] Check service destination is approved by NDID (SignData)

## 0.6.2 (August 8, 2018)

IMPROVEMENTS:

- [CircleCI] Update configuration for tendermint 0.22.8

BUG FIXES:

- [DeliverTx] Can not update RP and AS node (UpdateNodeByNDID)
- [DeliverTx] Check msq desination is not timed out (ClearRegisterMsqDestinationTimeout)

## 0.6.1 (August 8, 2018)

BUG FIXES:

- [Docker] Update path for download tendermint

## 0.6.0 (August 3, 2018)

BREAKING CHANGES:

- Change version of Tendermint to v0.22.8
- Change JSON property name `requestId` to `request_id` in parameter of all methods

IMPROVEMENTS:

- [DeliverTx] Check request is not closed (CloseRequest)
- [DeliverTx] Check request is not timed out (TimeOutRequest)
- [CheckTx] Validate public key (PEM format, RSA type with at least 2048-bit length is allowed)
- [Docker] Update Tendermint version to v0.22.8
- [Docker] Change Tendermint config to not create empty block (`create_empty_blocks = false`)
- [DeliverTx] Add valid_signature in response_valid_list to CloseRequest and TimeOutRequest parameter
- [DeliverTx] Add time out block of first MsqDestination (RegisterMsqDestination)
- [DeliverTx] Add function for set time out block (SetTimeOutBlockRegisterMsqDestination)

## 0.5.1 (July 22, 2018)

IMPROVEMENTS:

- Refactor code - Use switch-case instead of reflect pkg
- [DeliverTx] Remove check responseValid in CloseRequest and TimeOutRequest
- [Docker] Change Tendermint config to create empty block with interval of 30 seconds
- [DeliverTx] Add node_name to UpdateNodeByNDID parameter

BUG FIXES:

- [CheckTx] Return a correct code when receiving invalid transaction format
- [Query] Return a correct price when query "GetPriceFunc" with height

## 0.5.0 (July 16, 2018)

BREAKING CHANGES:

- Change version of Tendermint to v0.22.4

## 0.4.0 (July 14, 2018)

IMPROVEMENTS:

- [DeliverTx] Check responseValid in CloseRequest and TimeOutRequest

BREAKING CHANGES:

- [DeliverTx] Add master_public_key in parameter of InitNDID

## 0.3.0 (July 7, 2018)

FEATURES:

- [DeliverTx] Add new function (EnableMsqDestination, DisableMsqDestination, EnableAccessorMethod, DisableAccessorMethod, EnableService, DisableService, EnableNode, DisableNode, EnableNamespace, DisableNamespace, RegisterServiceDestinationByNDID, EnableServiceDestinationByNDID, DisableServiceDestinationByNDID, EnableServiceDestination, DisableServiceDestination)
- [CheckTx] Check method name

IMPROVEMENTS:

- [Docker] Use alpine:3.7 when building tendermint image

BREAKING CHANGES:

- Change version of Tendermint to v0.22.0
- [DeliverTx] Change transaction format
- [Query] Change query data format
- [DeliverTx] Before AS can RegisterServiceDestination need approval from NDID
- [DeliverTx] Change parameter of RegisterMsqDestination
- [Key/Value store] Add active flag in struct of MsqDestination, Accessor, Service
  , Node and Namespace
- [Query] Filter active flag (GetIdpNodes, GetAsNodesByServiceId, GetNamespaceList, GetServicesByAsID)

BUG FIXES:

- [DeliverTx] Fix missing `success` tag when creating a transaction with invalid signature

## 0.2.0 (June 30, 2018)

FEATURES:

- [CircleCI] Add a configuration for automatic test, build, and deploy image to dockerhub

BUG FIXES:

- [Query] Set special request if owner is IdP (GetRequestDetail)

## 0.1.1 (June 26, 2018)

BREAKING CHANGES:

- [Key/Value store] Remove key "NodePublicKeyRole"|<node’s public key> because allow to have duplicate key in network (unique only nodeID)

BUG FIXES:

- [DPKI] Fix when update public key with exist key in network and system set wrong role in stateDB

## 0.1.0 (June 24, 2018)

INITIAL:

- Initial release of NDID Smart Contract
