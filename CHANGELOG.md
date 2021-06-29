# Changelog

## TBD

IMPROVEMENTS:

- Add `lial` and `laal` to identity info.
- [DeliverTx] `RegisterIdentity` accepts `lial` and `laal`.
- [DeliverTx] `UpdateIdentity` accepts `lial` and `laal`.
- [DeliverTx] Make `ial` parameter optional for `UpdateIdentity`.
- [Query] Add `lial` and `laal` to result of `GetIdentityInfo`.

SECURITY FIXES:

- Disallow IdP nodes from creating request with data requests.
- Disallow IdP nodes from creating request with mode 1.

**IMPORTANT**: Security fixes mentioned above are considered breaking changes if deploy to the existing chain with transactions that violate the normal use cases (unexpected requests).

## 5.0.0 (May 7, 2020)

BREAKING CHANGES:

- Tendermint v0.33.3.
- [CheckTx / DeliverTx] Remove `SignData` method. Add `CreateAsResponse` method.

IMPROVEMENTS:

- Update Tendermint version to v0.33.3.
- [DeliverTx] `CreateIdpResponse` accepts `error_code`.
- [DeliverTx] `CreateAsResponse` accepts `error_code`.
- [DeliverTx] Add `AddErrorCode` method.
- [DeliverTx] Add `RemoveErrorCode` method.
- [DeliverTx] Add `whitelist` array and `agent` flag (only IdP) to node details.
- [Query] Add `agent` flag to IdP node details.
- [Query] Add `whitelist` array to node details.
- [Query] Add `agent` and `filter_for_node_id` parameters to `GetIdpNodes` and `GetIdpNodesInfo`.
- [Query] Add `whitelist` array and `agent` flag to `GetIdpNodes` and `GetIdpNodesInfo`.
- [Query] Add `GetErrorCodeList` method.
- [Docker] Use golang:1.14 when building image.
- [Docker] Use alpine:3.11 when building image.

## 4.1.0 (November 21, 2019)

IMPROVEMENTS:

- Save Tx signature check results in CheckTx and use them in DeliverTx - Attempt to reduce DeliverTx time and CPU consumption.
- Refactor app state, key name and prefixes.
- Change internal package name.
- [Docker] Change Go version used in images from 1.12 to 1.13.

## 4.0.0 (August 1, 2019)

BREAKING CHANGES:

- Tendermint v0.32.1.

IMPROVEMENTS:

- Use Go modules instead of dep.
- Update Tendermint version to v0.32.1.
- [Query] Add `active` property to result of `GetNodeInfo`.
- [Query] Add `ial` property to result of `GetIdpNodes` and `GetIdpNodesInfo`.
- [Docker] Update leveldb version to 1.22.
- [Docker] Remove default user.
- [Docker] Remove default owner and permission settings.
- [Docker] Remove `TERM` env.
- [Docker] Add docker-entrypoint.sh as image entrypoint which will check existence and owner of `ABCI_DB_DIR_PATH` and `TMHOME`.

OTHERS:

- [Docker] Remove `jq` and `curl` from docker image.

NOTES:

- [Docker] Docker container may be run with `-u` or `--user` flag (e.g. `-u 65534:65534`). In case you are using docker-compose, `user` may be specified in docker-compose file (e.g. `user: 65534:65534`) (see [Compose file reference](https://docs.docker.com/compose/compose-file/#domainname-hostname-ipc-mac_address-privileged-read_only-shm_size-stdin_open-tty-user-working_dir) for more detail).
- [Docker] When running docker container with non-root user, source directories that will be mounted into the container as `ABCI_DB_DIR_PATH` and `TMHOME` must be created beforehand with the non-root user as owner.

## 3.0.0 (May 29, 2019)

BREAKING CHANGES:

- [DeliverTx] Remove `first` and `hash_id` property from parameters of `RegisterIdentity`.
- [DeliverTx] Add `reference_group_code`, `new_identity_list`, `mode_list`, `accessor_id`, `accessor_public_key`, `accessor_type` and `request_id` property to parameters of `RegisterIdentity`.
- [DeliverTx] Remove `accessor_group_id` property from parameters of `AddAccessor`.
- [DeliverTx] Add `identity_namespace`, `identity_identifier_hash`, and `reference_group_code` property to parameters of `AddAccessor`.
- [DeliverTx] Change function name from `RevokeAccessorMethod` to `RevokeAccessor`.
- [DeliverTx] Add `supported_namespace_list` property to parameters of `RegisterServiceDestination`.
- [DeliverTx] Add `supported_namespace_list` property to parameters of `UpdateServiceDestination`.
- [DeliverTx] Remove `identity_proof` and `private_proof_hash` property from parameters of `CreateIdpResponse`.
- [DeliverTx] Remove `hash_id` property from parameters of `UpdateIdentity`.
- [DeliverTx] Add `reference_group_code`, `identity_namespace` and `identity_identifier_hash` property to parameters of `UpdateIdentity`.
- [DeliverTx] Add `allowed_identifier_count_in_reference_group` and `allowed_active_identifier_count_in_reference_group` property to parameters of `AddNamespace`.
- [DeliverTx] Remove `valid_proof` property from `response_valid_list` in parameters of `CloseRequest`.
- [DeliverTx] Remove `valid_proof` property from `response_valid_list` in parameters of `TimeOutRequest`.
- [DeliverTx] Add `supported_request_message_data_url_type_list` property to parameters of `UpdateNode` (IdP nodes only).
- [DeliverTx] Add new functions (`AddIdentity`, `UpdateIdentityModeList`, `RevokeIdentityAssociation`, `SetAllowedModeList`, `UpdateNamespace` and `SetAllowedMinIalForRegisterIdentityAtFirstIdp`).
- [DeliverTx] Remove `ClearRegisterIdentityTimeout`, `SetTimeOutBlockRegisterIdentity`, `RegisterAccessor` and `DeclareIdentityProof` function.
- [Query] Remove `hash_id` property from parameters of `GetIdpNodes`.
- [Query] Add `mode_list` and `supported_request_message_data_url_type_list` property to parameters of `GetIdpNodes`.
- [Query] Add `supported_request_message_data_url_type_list` property to list of `node` in result of `GetIdpNodes`.
- [Query] Remove `hash_id` property from parameters of `GetIdpNodesInfo`.
- [Query] Add `mode_list` and `supported_request_message_data_url_type_list` property to parameters of `GetIdpNodesInfo`.
- [Query] Add `supported_request_message_data_url_type_list` property to list of `node` in result of `GetIdpNodesInfo`.
- [Query] Add `supported_namespace_list` property to parameters of `GetAsNodesByServiceId`.
- [Query] Add `supported_namespace_list` property to parameters of `GetAsNodesInfoByServiceId`.
- [Query] Add `supported_namespace_list` property to parameters of `GetServicesByAsID`.
- [Query] Remove `hash_id` property from parameters of `CheckExistingIdentity`.
- [Query] Add `reference_group_code`, `identity_namespace` and `identity_identifier_hash` property to parameters of `CheckExistingIdentity`.
- [Query] Remove `hash_id` property from parameters of `GetIdentityInfo`.
- [Query] Add `reference_group_code`, `identity_namespace` and `identity_identifier_hash` property to parameters of `GetIdentityInfo`.
- [Query] Add `mode_list` property to result of `GetIdentityInfo`.
- [Query] Add `allowed_identifier_count_in_reference_group` and `allowed_active_identifier_count_in_reference_group` property to result of `GetNamespaceList`.
- [Query] Remove `identity_proof`, `private_proof_hash`, `valid_proof` from `response_list` in result of `GetIdentityInfo`.
- [Query] Add `supported_request_message_data_url_type_list` property to result of `GetNodeInfo` (IdP nodes only).
- [Query] Add `supported_request_message_data_url_type_list` property to result of `GetNodesBehindProxyNode` (IdP nodes only).
- [Query] Add new function (`GetReferenceGroupCode`, `GetReferenceGroupCodeByAccessorID`, `GetAllowedModeList` and `GetAllowedMinIalForRegisterIdentityAtFirstIdp`).
- [Query] Remove `GetIdentityProof` function.

IMPROVEMENTS:

- Update Tendermint version to v0.31.5.

## 2.0.0 (April 23, 2019)

BREAKING CHANGES:

- Change underlying data structure for app state storage.
- Remove IAVL dependency.
- Change DeliverTx logic for `recheck = false`.

IMPROVEMENTS:

- Update Tendermint version to v0.30.2.
- Add Prometheus support.

## 1.0.0 (December 7, 2018)

IMPROVEMENTS:

- Bundle ABCI app (proxy app) with Tendermint into single process.
- Use cleveldb instead of goleveldb.
- Update Tendermint version to v0.26.4.
- Change environment variable names and behavior
  - `DB_NAME` changed to `ABCI_DB_DIR_PATH`.
  - `LOG_LEVEL` changed to `ABCI_LOG_LEVEL`.
  - `LOG_TARGET` changed to `ABCI_LOG_TARGET` and accepts only either `console` or `file`.
  - New variable `ABCI_LOG_FILE_PATH` for specifying log file path when `ABCI_LOG_TARGET` is set to `file`.
  - New variable `ABCI_DB_TYPE` for specifying database backend type to use. Options are the same as Tendermint's `db_backend` config. Default is `cleveldb`.
- [Docker] Build Tendermint bundled with ABCI app (proxy app) image with cleveldb support.
- [Migrate script] Move script to `migration-tools` repo.

SECURITY FIXES:

- [DeliverTx] Fix `nonce` does not get saved to state DB in every transactions.

BUG FIXES:

- [DeliverTx] Return invalid error when set non existent validator (`SetValidator`).
- [DeliverTx] Check request is not closed or timed out (`SetDataReceived`).

## 0.13.0 (November 21, 2018)

BREAKING CHANGES:

- Change version of Tendermint to v0.26.3

IMPROVEMENTS:

- [Docker] Add set config `recheck` to true to start script. (Fix transactions stuck in mempool until more Tx is broadcasted in some cases when there are multiple validator nodes.)
- [Info] Add app protocol version.

BUG FIXES:

- [Migrate script] Filter out `setLastBlock` key.

## 0.12.0 (November 15, 2018)

IMPROVEMENTS:

- [DeliverTx] Add `chain_history_info` property to parameters of `InitNDID`.
- [Query] Add new function (`GetChainHistory`).

SECURITY FIXES:

- [CheckTx] Check duplicate `nonce` in every transactions.

## 0.11.2 (November 12, 2018)

IMPROVEMENTS:

- [Docker] Add set config `recheck` and `recheck_empty` to false to start script.

BUG FIXES:

- [CheckTx] Add previously removed (in v0.11.1) transaction checking.

## 0.11.1 (November 11, 2018)

BREAKING CHANGES:

- Use deterministic Protobuf marshalling.
- Use Protobuf to store node's token instead of string.
- Remove Tx price history recording on DeliverTx since it causes performance issue when there are a lot of stored token usage records.
- Remove `GetUsedTokenReport` function.
- [DeliverTx] Check node ID is valid and role is `AS` (`RegisterServiceDestinationByNDID`, `EnableServiceDestinationByNDID` and `DisableServiceDestinationByNDID`).
- [DeliverTx] Add new functions (`RevokeAccessorMethod`).
- [DeliverTx] Check request is closed with valid ial, proof and signature (`AddAccessorMethod`).
- [DeliverTx] Add new functions (`SetLastBlock`) for disble create transaction.

IMPROVEMENTS:

- [Query] Add new function (`GetAccessorOwner`).
- [Dependency] Update iavl version to 0.11.0.
- [Query] Add `creation_chain_id` property to result of `GetRequestDetail`.
- [DeliverTx] Add new function (`EndInit`) for NDID.

## 0.10.2 (October 29, 2018)

NOTHING CHANGES, BUMP VERSION TO MATCH API REPOSITORY

## 0.10.1 (October 9, 2018)

IMPROVEMENTS:

- Align the happy path to the left edge.

## 0.10.0 (October 7, 2018)

BREAKING CHANGES:

- [DeliverTx] Check all IdP in `idp_id_list` is active (`CreateRequest`).
- [DeliverTx] Check all AS in `as_id_list` is active (`CreateRequest`).

IMPROVEMENTS:

- [Query] Filter proxy node is not active `GetIdpNodesInfo`.
- [Query] Filter proxy node is not active `GetAsNodesInfoByServiceId`.

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
