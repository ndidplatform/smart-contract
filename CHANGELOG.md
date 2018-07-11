# Changelog

## X.X.X

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
