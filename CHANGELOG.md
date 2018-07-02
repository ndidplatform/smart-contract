# Changelog

## 0.2.0 (June 30, 2018)

FEATURES:
- [CircleCI] Add a configuration for automatic test, build, and deploy image to dockerhub.

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
