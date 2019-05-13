# ABCI v2.0

## RegisterIdentity

### Parameter

```json
{
  "reference_group_code": "aaaaa-bbbbb-ccccc-ddddd",
  "new_identity_list": [{
    "identity_namespace": "citizenId",
    "identity_identifier_hash": "c765a80f1ee71299c361c1b4cb4d9c36b44061a526348a71287ea0a97cea80f6",
  }],
  "ial": 3,
  "mode_list": [2, 3], // allow only 2, 3
  "accessor_id": "11267a29-2196-4400-8b67-7424519b87ec",
  "accessor_public_key": "-----BEGIN PUBLIC KEY-----\\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA7BjIuleY9/5ObFl0w+U2\\nfID4cC8v3yIaOjsImXYNon04TZ6lHs8gNvrR1Q0MRtGTugL8XJPj3tw1AbHj01L8\\nW0HwKpFQxhwvGzi0Sesb9Lhn9aA4MCmfMG7PwLGzgdeHR7TVl7VhKx7gedyYIdju\\nEFzAtsJYO1plhUfFv6gdg/05VOjFTtVdWtwKgjUesmuv1ieZDj64krDS84Hka0gM\\njNKm4+mX8HGUPEkHUziyBpD3MwAzyA+I+Z90khDBox/+p+DmlXuzMNTHKE6bwesD\\n9ro1+LVKqjR/GjSZDoxL13c+Va2a9Dvd2zUoSVcDwNJzSJtBrxMT/yoNhlUjqlU0\\nYQIDAQAB\\n-----END PUBLIC KEY-----",
  "accessor_type": "accessor_type",
  "request_id": "edaec8df-7865-4473-8707-054dd0cffe2d"
}
```

**NOTE**

- Remove `first` property
- Remove `hash_id`
- Add `reference_group_code` property (string)
- Add all identity in `new_identity_list`.
- Add `mode_list`
- Add `accessor_id`
- Add `accessor_public_key`
- Add `accessor_type`
- Add `request_id`
- All identity in `new_identity_list` MUST not exist.

## AddIdentity

### Parameter

```json
{
  "reference_group_code": "aaaaa-bbbbb-ccccc-ddddd",
  "new_identity_list": [{
    "identity_namespace": "citizenId",
    "identity_identifier_hash": "c765a80f1ee71299c361c1b4cb4d9c36b44061a526348a71287ea0a97cea80f6",
  }],
  "request_id": "edaec8df-7865-4473-8707-054dd0cffe2d"
}
```

**NOTE**

- `reference_group_code` MUST already exist.
- All identity in `new_identity_list` MUST not exist.
- IdP must already be associated with `reference_group_code`

## AddAccessor

### Parameter

```json
{
  "reference_group_code": "aaaaa-bbbbb-ccccc-ddddd",
  "identity_namespace": "citizenId",
  "identity_identifier_hash": "c765a80f1ee71299c361c1b4cb4d9c36b44061a526348a71287ea0a97cea80f6",
  "accessor_id": "07938aa2-2aaf-4bb5-9ccd-33700581e870",
  "accessor_public_key": "-----BEGIN PUBLIC KEY-----\\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAhdKdvawPO8XXroiAGkxF\\nfLRCqvk4X2iAMStq1ADjmPPWhKgF/ssU9LBdHKHPPX1+NMOX29gOL3ZCxfZamKO6\\nAbODt1e0bVfblWWMq5uMwzNrFo4nKas74SLJwiMg0vtn1NnHU4QTTrMYmGqRf2WZ\\nIN9Iro4LytUTLEBCpimWM2hodO8I60bANAO0gI96BzAWMleoioOzWlq6JKkiDsj7\\n8EjCI/bY1T/v4F7rg2FxrIH/BH4TUDy88pIvAYy4nNEyGyr8KzMm1cKxOgnJI8On\\nwT8HrAJQ58T3HCCiCrKAohkYBWITPk3cmqGfOKrqZ2DI+a6URofMVvQFlwfYvqU6\\n5QIDAQAB\\n-----END PUBLIC KEY-----",
  "accessor_type": "accessor_type_2",
  "request_id": "edaec8df-7865-4473-8707-054dd0cffe2d"
}
```

**NOTE**

- Remove `accessor_group_id`
- Add `identity_namespace`, `identity_identifier_hash`, and `reference_group_code`
- Input `reference_group_code` or `identity_namespace`+`identity_identifier_hash` (able to input one or the other, if both then error)
- Change function name from `AddAccessorMethod` to `AddAccessor`


## UpdateIdentityModeList (New)

### Parameter

```json
{
  "reference_group_code": "aaaaa-bbbbb-ccccc-ddddd",
  "identity_namespace": "citizenId",
  "identity_identifier_hash": "c765a80f1ee71299c361c1b4cb4d9c36b44061a526348a71287ea0a97cea80f6",
  "mode_list": [2, 3], // allow only 2,3
  "request_id": "edaec8df-7865-4473-8707-054dd0cffe2d"
}
```

**NOTE**

- If identity is already in mode that's in input parameter, error.
- If mode in input is less than identity's current mode, error.
- Input `reference_group_code` or `identity_namespace`+`identity_identifier_hash` (able to input one or the other, if both then error)


## GetReferenceGroupCode

### Parameter

```json
{
  "identity_namespace": "citizenId",
  "identity_identifier_hash": "c765a80f1ee71299c361c1b4cb4d9c36b44061a526348a71287ea0a97cea80f6",
}
```

### Output

```json
{
  "reference_group_code": "aaaaa-bbbbb-ccccc-ddddd"
}
```


## GetReferenceGroupCodeByAccessorID

### Parameter

```json
{
  "accessor_id": "11267a29-2196-4400-8b67-7424519b87ec",
}
```

### Output

```json
{
  "reference_group_code": "aaaaa-bbbbb-ccccc-ddddd"
}
```


## GetIdpNodes

### Parameter

```json
{
  "reference_group_code": "aaaaa-bbbbb-ccccc-ddddd",
  "identity_namespace": "citizenId",
  "identity_identifier_hash": "c765a80f1ee71299c361c1b4cb4d9c36b44061a526348a71287ea0a97cea80f6",
  "min_aal": 3,
  "min_ial": 3,
  "node_id_list": [], //array of string
  "supported_request_message_data_url_type_list": [], //array of string
  "mode_list": [3]
}
```

### Output

```json
{
  "node": [
    {
      "max_aal": 3,
      "max_ial": 3,
      "node_id": "CuQfyyhjGcCAzKREzHmL",
      "node_name": "IdP Number 1 from ...",
      "mode_list": [2, 3], //array of available mode
      "supported_request_message_data_url_type_list": ["text/plain", "application/pdf"]
    }
  ]
}
```

**NOTE**

- Input `reference_group_code` or `identity_namespace`+`identity_identifier_hash` (able to input one or the other, if both then error when `identity_namespace`+`identity_identifier_hash` is not in that reference_group_code?)
- Remove `hash_id`
- Add `mode_list`
- Add `supported_request_message_data_url_type_list`


## GetIdpNodesInfo

### Parameter

```json
{
  "reference_group_code": "aaaaa-bbbbb-ccccc-ddddd",
  "identity_namespace": "citizenId",
  "identity_identifier_hash": "c765a80f1ee71299c361c1b4cb4d9c36b44061a526348a71287ea0a97cea80f6",
  "min_aal": 3,
  "min_ial": 3,
  "node_id_list": [], //array of string
  "supported_request_message_data_url_type_list": [], //array of string
  "mode_list": [3]
}
```

### Expected Output

```json
{
  "node": [
    {
      "max_aal": 3,
      "max_ial": 3,
      "mode_list": [2, 3], //array of available mode
      "supported_request_message_data_url_type_list": ["text/plain", "application/pdf"],
      "mq": [
        {
          "ip": "192.168.3.99",
          "port": 8000
        }
      ],
      "name": "IdP Number 1 from ...",
      "node_id": "CuQfyyhjGcCAzKREzHmL",
      "public_key": "-----BEGIN PUBLIC KEY-----\\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\\njwIDAQAB\\n-----END PUBLIC KEY-----\\n"
    }
  ]
}
```

**NOTE**

- Input `reference_group_code` or `identity_namespace`+`identity_identifier_hash` (able to input one or the other, if both then error when `identity_namespace`+`identity_identifier_hash` is not in that reference_group_code?)
- Remove `hash_id`
- Add `mode_list`
- Add `supported_request_message_data_url_type_list`


## RevokeAccessor

### Parameter

```json
{
  "accessor_id_list": [
    "11d10976-aede-4ba0-9f44-fc0c96db1f32"
  ],
  "request_id": "e7dcf1c2-eea7-4dc8-af75-724cf86454ef"
}
```

**NOTE**

- Change function name from `RevokeAccessorMethod` to `RevokeAccessor`


## RevokeIdentityAssociation (New)

### Parameter

```json
{
  "reference_group_code": "aaaaa-bbbbb-ccccc-ddddd",
  "identity_namespace": "citizenId",
  "identity_identifier_hash": "c765a80f1ee71299c361c1b4cb4d9c36b44061a526348a71287ea0a97cea80f6",
  "request_id": "edaec8df-7865-4473-8707-054dd0cffe2d"
}
```

## MergeReferenceGroup (New)

### Parameter

```json
{
  "reference_group_code": "aaaaa-bbbbb-ccccc-ddddd",
  "identity_namespace": "citizenId",
  "identity_identifier_hash": "c765a80f1ee71299c361c1b4cb4d9c36b44061a526348a71287ea0a97cea80f6",
  "reference_group_code_to_merge": "eeeee-fffff-ggggg-hhhhh", // Merge to `reference_group_code`
  "identity_namespace_to_merge": "citizenId",
  "identity_identifier_hash_to_merge": "c765a80f1ee71299c361c1b4cb4d9c36b44061a526348a71287ea0a97cea80f6",
  "request_id": "edaec8df-7865-4473-8707-054dd0cffe2d"
}
```

**NOTE**

- Input `reference_group_code` or `identity_namespace`+`identity_identifier_hash` (able to input one or the other, if both, error)
- Input `reference_group_code_to_merge` or `identity_namespace_to_merge`+`identity_identifier_hash_to_merge` (able to input one or the other, if both, error)
- If `reference_group_code` == `reference_group_code_to_merge`, error
- If `identity_namespace` == `identity_namespace_to_merge` && `identity_identifier_hash` == `identity_identifier_hash_to_merge`, error
- Mark old `reference_group_code` that it's merged (effectively disable)


## RegisterServiceDestination

### Parameter

```json
{
  "min_aal": 1.2,
  "min_ial": 1.1,
  "service_id": "LlUXaAYeAoVDiQziKPMc",
  "supported_namespace_list": [
    "citizenId"
  ]
}
```

**NOTE**

- Add `supported_namespace_list`


## UpdateServiceDestination

### Parameter

```json
{
  "min_aal": 1.5,
  "min_ial": 1.4,
  "service_id": "LlUXaAYeAoVDiQziKPMc",
  "supported_namespace_list": [
    "citizenId"
  ]
}
```

**NOTE**

- Add `supported_namespace_list`


## GetAsNodesByServiceId

### Expected Output

```json
{
  "node": [
    {
      "min_aal": 1.5,
      "min_ial": 1.4,
      "node_id": "XckRuCmVliLThncSTnfG",
      "node_name": "AS1",
      "supported_namespace_list": [
        "citizenId"
      ]
    }
  ]
}
```

**NOTE**

- Add `supported_namespace_list`


## GetAsNodesInfoByServiceId

### Expected Output

```json
{
  "node": [
    {
      "min_aal": 1.5,
      "min_ial": 1.4,
      "mq": [
        {
          "ip": "192.168.3.102",
          "port": 8000
        }
      ],
      "name": "AS1",
      "node_id": "XckRuCmVliLThncSTnfG",
      "public_key": "-----BEGIN PUBLIC KEY-----\\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEApT8lXT9CDRZZkvhZLBD6\\n6o7igZf6sj/o0XooaTuy2HuCt6yEO8jt7nx0XkEFyx4bH4/tZNsKdok7DU75MjqQ\\nrdqGwpogvkZ3uUahwE9ZgOj6h4fq9l1Au8lxvAIp+b2BDRxttbHp9Ls9nK47B3Zu\\niD02QknUNiPFvf+BWIoC8oe6AbyctnV+GTsC/H3jY3BD9ox2XKSE4/xaDMgC+SBU\\n3pqukT35tgOcvcSAMVJJ06B3uyk19MzK3MVMm8b4sHFQ76UEpDOtQZrmKR1PH0gV\\nFt93/0FPOH3m4o+9+1OStP51Un4oH3o80aw5g0EJzDpuv/+Sheec4+0PVTq0K6kj\\ndQIDAQAB\\n-----END PUBLIC KEY-----\\n",
      "supported_namespace_list": [
        "citizenId"
      ]
    }
  ]
}
```

**NOTE**

- Add `supported_namespace_list`


## GetServicesByAsID

### Expected Output

```json
{
  "services": [
    {
      "active": true,
      "min_aal": 1.1,
      "min_ial": 1.1,
      "service_id": "AFLHeKQVLNQOkIOxoNid",
      "supported_namespace_list": [
        "citizenId"
      ],
      "suspended": false
    },
    {
      "active": true,
      "min_aal": 2.2,
      "min_ial": 2.2,
      "service_id": "qvyfrfJRsfaesnDsYHbH",
      "supported_namespace_list": [
        "citizenId"
      ],
      "suspended": false
    },
    {
      "active": true,
      "min_aal": 3.3,
      "min_ial": 3.3,
      "service_id": "JTFHqDoJRccWcikcJqnL",
      "supported_namespace_list": [
        "citizenId"
      ],
      "suspended": false
    }
  ]
}
```

**NOTE**

- Add `supported_namespace_list`


## CreateIdpResponse

### Parameter

```json
{
  "aal": 3,
  "ial": 3,
  "request_id": "16dc0550-a6e4-4e1f-8338-37c2ac85af74",
  "signature": "signature",
  "status": "accept"
}
```

**NOTE**

- Remove `identity_proof`
- Remove `private_proof_hash`


## CheckExistingIdentity

### Parameter

```json
{
  "reference_group_code": "aaaaa-bbbbb-ccccc-ddddd",
  "identity_namespace": "citizenId",
  "identity_identifier_hash": "c765a80f1ee71299c361c1b4cb4d9c36b44061a526348a71287ea0a97cea80f6",
}
```

**NOTE**

- Remove `hash_id`
- Add `reference_group_code`, `identity_namespace` and `identity_identifier_hash`
- Input `reference_group_code` or `identity_namespace`+`identity_identifier_hash` (able to input one or the other, if both, error)


## GetIdentityInfo

### Parameter

```json
{
  "reference_group_code": "aaaaa-bbbbb-ccccc-ddddd",
  "identity_namespace": "citizenId",
  "identity_identifier_hash": "c765a80f1ee71299c361c1b4cb4d9c36b44061a526348a71287ea0a97cea80f6",
  "node_id": "CuQfyyhjGcCAzKREzHmL"
}
```

### Expected Output

```json
{
  "ial": 2.2,
  "mode_list": [2, 3]
}
```

**NOTE**

- Remove `hash_id`
- Add `reference_group_code`, `identity_namespace` and `identity_identifier_hash`
- Input `reference_group_code` or `identity_namespace`+`identity_identifier_hash` (able to input one or the other, if both, error)
- Add `mode_list` to output


## UpdateIdentity

### Parameter

```json
{
  "reference_group_code": "aaaaa-bbbbb-ccccc-ddddd",
  "identity_namespace": "citizenId",
  "identity_identifier_hash": "c765a80f1ee71299c361c1b4cb4d9c36b44061a526348a71287ea0a97cea80f6",
  "ial": 2.2
}
```

**NOTE**

- Remove `hash_id`
- Add `reference_group_code`, `identity_namespace` and `identity_identifier_hash`
- Input `reference_group_code` or `identity_namespace`+`identity_identifier_hash` (able to input one or the other, if both, error)


## GetAllowedModeList

### Parameter

```json
{
  "purpose": "AddAccessor",
}
```

### Expected Output
```json
{
  "allowed_mode_list": [1, 2, 3]
}
```

## SetAllowedModeList

### Parameter

```json
{
  "purpose": "",
  "allowed_mode_list": [1, 2, 3],
}
```

**NOTE**

- Only request with mode in `allowed_mode_list` can be created.
- Only NDID can update `allowed_mode_list`.
- Purpose for normal transaction is empty string.


## UpdateNamespace

### Parameter

```json
{
  "description": "Citizen ID",
  "namespace": "citizenId",
  "allowed_identifier_count_in_reference_group": 1,
  "allowed_active_identifier_count_in_reference_group": 1
}
```

**NOTE**

- Only NDID can call this function
- if `allowed_identifier_count_in_reference_group` is not present or 0 or -1, means unlimited
- if `allowed_active_identifier_count_in_reference_group` is not present or 0 or -1, means unlimited

## AddNamespace

### Parameter

```json
{
  "description": "Citizen ID",
  "namespace": "WsvGOEjoFqvXsvcfFVWm",
  "allowed_identifier_count_in_reference_group": 1,
  "allowed_active_identifier_count_in_reference_group": 1
}
```

**NOTE**

- Only NDID can call this function
- if `allowed_identifier_count_in_reference_group` is not present or 0 or -1, means unlimited
- if `allowed_active_identifier_count_in_reference_group` is not present or 0 or -1, means unlimited

### Expected Output

```json
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## GetNamespaceList

### Expected Output

```json
[
  {
    "namespace": "WsvGOEjoFqvXsvcfFVWm",
    "description": "Citizen ID",
    "active": true,
    "allowed_identifier_count_in_reference_group": 1,
    "allowed_active_identifier_count_in_reference_group": 1
  },
  {
    "namespace": "SJsMIeJcerfZpBfXkJgU",
    "description": "Tel number",
    "active": true,
  }
]
```
**NOTE**

- if `allowed_identifier_count_in_reference_group` is not present, means unlimited
- if `allowed_active_identifier_count_in_reference_group` is not present, means unlimited

## SetAllowedMinIalForRegisterIdentityAtFirstIdp

### Parameter

```json
{
  "min_ial": 2.3,
}
```

## GetAllowedMinIalForRegisterIdentityAtFirstIdp

### Expected Output

```json
{
  "min_ial": 2.3,
}
```

## CloseRequest

### Parameter

```json
{
  "request_id": "16dc0550-a6e4-4e1f-8338-37c2ac85af74",
  "response_valid_list": [
    {
      "idp_id": "CuQfyyhjGcCAzKREzHmL",
      "valid_ial": true,
      "valid_signature": true
    }
  ]
}
```

**NOTE**

- Remove `valid_proof`

## TimeOutRequest

### Parameter

```json
{
  "request_id": "04db0ddf-4d3f-4b40-93b0-af418ad8a2d7",
  "response_valid_list": [
    {
      "idp_id": "CuQfyyhjGcCAzKREzHmL",
      "valid_ial": false,
      "valid_signature": false
    }
  ]
}
```

**NOTE**

- Remove `valid_proof`


## GetRequestDetail

### Expected Output

```json
{
  "closed": false,
  "data_request_list": [
    {
      "answered_as_id_list": [
        "XckRuCmVliLThncSTnfG"
      ],
      "as_id_list": [],
      "min_as": 1,
      "received_data_from_list": [
        "XckRuCmVliLThncSTnfG"
      ],
      "request_params_hash": "hash",
      "service_id": "LlUXaAYeAoVDiQziKPMc"
    }
  ],
  "min_aal": 3,
  "min_ial": 3,
  "min_idp": 1,
  "mode": 3,
  "request_id": "16dc0550-a6e4-4e1f-8338-37c2ac85af74",
  "request_message_hash": "hash('Please allow...')",
  "request_timeout": 259200,
  "idp_id_list": [
    "lvEzsuTcZvIRvZyrdEsi",
    "njHtYuHHxCvzzofcpwon"
  ],
  "requester_node_id": "nfhwDGTTeRdMeXzAgLij",
  "response_list": [
    {
      "aal": 3,
      "ial": 3,
      "idp_id": "CuQfyyhjGcCAzKREzHmL",
      "signature": "signature",
      "status": "accept",
      "valid_ial": null,
      "valid_signature": null
    }
  ],
  "purpose": "",
  "timed_out": false,
  "creation_block_height": 50,
  "creation_chain_id": "test-chain-NDID"
}
```

**NOTE**

- Remove `identity_proof`, `private_proof_hash`, `valid_proof` from `response_list`


## UpdateNode

### Parameter

```json
{
  "master_public_key": "",
  "public_key": "-----BEGIN PUBLIC KEY-----\\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArdcKj/gAetVyg6Nn2lDi\\nm/UJYQsQCav60EVbECm5EVT8WgnpzO+GrRyBtxqWUdtGar7d6orLh1RX1ikU7Yx2\\nSA8Xlf+ZDaCELba/85Nb+IppLBdPywixgumoto9G9dDGSnPkHAlq5lXXA1eeUS7j\\niU1lf37lwTZaO0COAuu8Vt9GcwYPh7SSf4/eXabQGbo/TMUVpXX1w5N1A07Qh5DG\\nr/ZKzEE9/5bJJJRS635OA2T4gIY9XRWYiTxtiZz6AFCxP92Cjz/sNvSc/Cuvwi15\\nycS4C35tjM8iT5djsRcR+MJeXyvurkaYgMGJTDIWub/A5oavVD3VwusZZNZvpDpD\\nPwIDAQAB\\n-----END PUBLIC KEY-----\\n",
  "supported_request_message_data_url_type_list": ["text/plain", "application/pdf"] // For IdP node only, "text/plain" must be supported always
}
```

**NOTE**

- Add `supported_request_message_data_url_type_list` (IdP nodes only)


## GetNodeInfo

### Expected Output

```json
{
  "master_public_key": "-----BEGIN PUBLIC KEY-----\\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAukTxVg8qpwXebALGCrly\\niv8PNNxLo0CEX3N33cR1TNfImItd5nFwmozLJLM9LpNF711PrkH3EBLJM+qwASlC\\nBayeMiMT8tDmOtv1RqIxyLjEU8M0RBBedk/TsKQwNmmeU3n5Ap+GRTYoEOwTKNra\\nI8YDfbjb9fNtSICiDzn3UcQj13iLz5x4MjaewtC6PR1r8uVfLyS4uI+3/qau0zWV\\n+s6b3JdqU2zdHeuaj9XjX7aNV7mvnjYgzk/O7M/p/86RBEOm7pt6JmTGnFu44jBO\\nez6GqF2hZzqR9nM1K4aOedBMHintVnhh1oOPG9uRiDnJWvN16PNTfr7XBOUzL03X\\nDQIDAQAB\\n-----END PUBLIC KEY-----\\n",
  "max_aal": 2.4,
  "max_ial": 2.3,
  "supported_request_message_data_url_type_list": ["text/plain", "application/pdf"],
  "mq": [
    {
      "ip": "192.168.3.99",
      "port": 8000
    }
  ],
  "node_name": "IdP Number 1 from ...",
  "public_key": "-----BEGIN PUBLIC KEY-----\\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArdcKj/gAetVyg6Nn2lDi\\nm/UJYQsQCav60EVbECm5EVT8WgnpzO+GrRyBtxqWUdtGar7d6orLh1RX1ikU7Yx2\\nSA8Xlf+ZDaCELba/85Nb+IppLBdPywixgumoto9G9dDGSnPkHAlq5lXXA1eeUS7j\\niU1lf37lwTZaO0COAuu8Vt9GcwYPh7SSf4/eXabQGbo/TMUVpXX1w5N1A07Qh5DG\\nr/ZKzEE9/5bJJJRS635OA2T4gIY9XRWYiTxtiZz6AFCxP92Cjz/sNvSc/Cuvwi15\\nycS4C35tjM8iT5djsRcR+MJeXyvurkaYgMGJTDIWub/A5oavVD3VwusZZNZvpDpD\\nPwIDAQAB\\n-----END PUBLIC KEY-----\\n",
  "role": "IdP"
}
```

**NOTE**

- Add `supported_request_message_data_url_type_list` to IdP nodes


## GetNodesBehindProxyNode

### Expected Output

```json
{
  "nodes": [
    {
      "config": "KEY_ON_PROXY",
      "master_public_key": "-----BEGIN PUBLIC KEY-----\\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\\njwIDAQAB\\n-----END PUBLIC KEY-----\\n",
      "max_aal": 3,
      "max_ial": 3,
      "supported_request_message_data_url_type_list": ["text/plain", "application/pdf"],
      "node_id": "BLUbbuoywxSirpxDIPgW",
      "node_name": "IdP6BehindProxy1",
      "public_key": "-----BEGIN PUBLIC KEY-----\\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\\njwIDAQAB\\n-----END PUBLIC KEY-----\\n",
      "role": "IdP"
    }
  ]
}
```

**NOTE**

- Add `supported_request_message_data_url_type_list` to IdP nodes

## RevokeAndAddAccessor

### Parameter

```json
{
  "revoke_accessor_id_list": [
    "11d10976-aede-4ba0-9f44-fc0c96db1f32"
  ],
  "reference_group_code": "aaaaa-bbbbb-ccccc-ddddd",
  "identity_namespace": "citizenId",
  "identity_identifier_hash": "c765a80f1ee71299c361c1b4cb4d9c36b44061a526348a71287ea0a97cea80f6",
  "accessor_id": "07938aa2-2aaf-4bb5-9ccd-33700581e870",
  "accessor_public_key": "-----BEGIN PUBLIC KEY-----\\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAhdKdvawPO8XXroiAGkxF\\nfLRCqvk4X2iAMStq1ADjmPPWhKgF/ssU9LBdHKHPPX1+NMOX29gOL3ZCxfZamKO6\\nAbODt1e0bVfblWWMq5uMwzNrFo4nKas74SLJwiMg0vtn1NnHU4QTTrMYmGqRf2WZ\\nIN9Iro4LytUTLEBCpimWM2hodO8I60bANAO0gI96BzAWMleoioOzWlq6JKkiDsj7\\n8EjCI/bY1T/v4F7rg2FxrIH/BH4TUDy88pIvAYy4nNEyGyr8KzMm1cKxOgnJI8On\\nwT8HrAJQ58T3HCCiCrKAohkYBWITPk3cmqGfOKrqZ2DI+a6URofMVvQFlwfYvqU6\\n5QIDAQAB\\n-----END PUBLIC KEY-----",
  "accessor_type": "accessor_type_2",
  "request_id": "edaec8df-7865-4473-8707-054dd0cffe2d"
}
```

## Remove these functions

- ClearRegisterIdentityTimeout 
- SetTimeOutBlockRegisterIdentity
- RegisterAccessor
- DeclareIdentityProof
- GetIdentityProof
