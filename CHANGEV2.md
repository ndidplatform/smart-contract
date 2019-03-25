# ABCI v2.0

## RegisterIdentity

### Parameter

```javascript
{
  "users": [
    {
      "reference_group_code": "aaaaa-bbbbb-ccccc-ddddd",
      "identity_namespace": "citizenId",
      "identity_identifier_hash": "c765a80f1ee71299c361c1b4cb4d9c36b44061a526348a71287ea0a97cea80f6",
      "ial": 3,
      "mode": [2, 3], // allow only 2, 3
      "accessor_id": "11267a29-2196-4400-8b67-7424519b87ec",
      "accessor_public_key": "-----BEGIN PUBLIC KEY-----\\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA7BjIuleY9/5ObFl0w+U2\\nfID4cC8v3yIaOjsImXYNon04TZ6lHs8gNvrR1Q0MRtGTugL8XJPj3tw1AbHj01L8\\nW0HwKpFQxhwvGzi0Sesb9Lhn9aA4MCmfMG7PwLGzgdeHR7TVl7VhKx7gedyYIdju\\nEFzAtsJYO1plhUfFv6gdg/05VOjFTtVdWtwKgjUesmuv1ieZDj64krDS84Hka0gM\\njNKm4+mX8HGUPEkHUziyBpD3MwAzyA+I+Z90khDBox/+p+DmlXuzMNTHKE6bwesD\\n9ro1+LVKqjR/GjSZDoxL13c+Va2a9Dvd2zUoSVcDwNJzSJtBrxMT/yoNhlUjqlU0\\nYQIDAQAB\\n-----END PUBLIC KEY-----",
      "accessor_type": "accessor_type",
      "request_id": "edaec8df-7865-4473-8707-054dd0cffe2d"
    }
  ]
}
```

**NOTE**

- Remove `first` property
- Add `reference_group_code` property (string)
- Remove `hash_id`
- Add `identity_namespace`
- Add `identity_identifier_hash`
- Add `mode`
- Add `accessor_id`
- Add `accessor_public_key`
- Add `accessor_type`
- Add `request_id`
- Check for `identity_namespace`+`identity_identifier_hash`. If exist, error.


## AddAccessorMethod

### Parameter

```javascript
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


## UpdateIdentityModeList (New)

### Parameter

```javascript
{
  "reference_group_code": "aaaaa-bbbbb-ccccc-ddddd",
  "identity_namespace": "citizenId",
  "identity_identifier_hash": "c765a80f1ee71299c361c1b4cb4d9c36b44061a526348a71287ea0a97cea80f6",
  "mode": [2, 3], // allow only 2,3
  "request_id": "edaec8df-7865-4473-8707-054dd0cffe2d"
}
```

**NOTE**

- If identity is already in mode that's in input parameter, error.
- If mode in input is less than identity's current mode, error.
- Input `reference_group_code` or `identity_namespace`+`identity_identifier_hash` (able to input one or the other, if both then error)


## GetReferenceGroupCode

### Parameter

```javascript
{
  "identity_namespace": "citizenId",
  "identity_identifier_hash": "c765a80f1ee71299c361c1b4cb4d9c36b44061a526348a71287ea0a97cea80f6",
}
```

### Output

```javascript
{
  "reference_group_code": "aaaaa-bbbbb-ccccc-ddddd"
}
```


## GetReferenceGroupCodeByAccessorId

### Parameter

```javascript
{
  "accessor_id": "11267a29-2196-4400-8b67-7424519b87ec",
}
```

### Output

```javascript
{
  "reference_group_code": "aaaaa-bbbbb-ccccc-ddddd"
}
```


## GetIdpNodes

### Parameter

```javascript
{
  "reference_group_code": "aaaaa-bbbbb-ccccc-ddddd",
  "identity_namespace": "citizenId",
  "identity_identifier_hash": "c765a80f1ee71299c361c1b4cb4d9c36b44061a526348a71287ea0a97cea80f6",
  "min_aal": 3,
  "min_ial": 3,
  "node_id_list": [] //array of string
}
```

### Output

```javascript
{
  "node": [
    {
      "max_aal": 3,
      "max_ial": 3,
      "node_id": "CuQfyyhjGcCAzKREzHmL",
      "node_name": "IdP Number 1 from ...",
      "mode": [2, 3] //array of available mode
    }
  ]
}
```

**NOTE**

- Input `reference_group_code` or `identity_namespace`+`identity_identifier_hash` (able to input one or the other, if both then error when `identity_namespace`+`identity_identifier_hash` is not in that reference_group_code?)
- Remove `hash_id`
- Add `mode` to result of `GetIdpNodesInfo`


## RevokeAccessorMethod

### Parameter

```javascript
{
  "accessor_id_list": [
    "11d10976-aede-4ba0-9f44-fc0c96db1f32"
  ],
  "request_id": "e7dcf1c2-eea7-4dc8-af75-724cf86454ef"
}
```


## RevokeIdentity (New)

### Parameter

```javascript
{
  "reference_group_code": "aaaaa-bbbbb-ccccc-ddddd",
  "identity_namespace": "citizenId",
  "identity_identifier_hash": "c765a80f1ee71299c361c1b4cb4d9c36b44061a526348a71287ea0a97cea80f6",
  "request_id": "edaec8df-7865-4473-8707-054dd0cffe2d"
}
```


## MergeReferenceGroup (New)

### Parameter

```javascript
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


## RegisterServiceDestination

### Parameter

```sh
{
  "min_aal": 1.2,
  "min_ial": 1.1,
  "service_id": "LlUXaAYeAoVDiQziKPMc",
  "accepted_namespace_list": [
    "citizenId"
  ]
}
```

**NOTE**

- Add `accepted_namespace_list`


## UpdateServiceDestination

### Parameter

```sh
{
  "min_aal": 1.5,
  "min_ial": 1.4,
  "service_id": "LlUXaAYeAoVDiQziKPMc",
  "accepted_namespace_list": [
    "citizenId"
  ]
}
```

**NOTE**

- Add `accepted_namespace_list`


## GetAsNodesByServiceId

### Expected Output

```sh
{
  "node": [
    {
      "min_aal": 1.5,
      "min_ial": 1.4,
      "node_id": "XckRuCmVliLThncSTnfG",
      "node_name": "AS1",
      "accepted_namespace_list": [
        "citizenId"
      ]
    }
  ]
}
```

**NOTE**

- Add `accepted_namespace_list`


## GetAsNodesInfoByServiceId

### Expected Output

```sh
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
      "accepted_namespace_list": [
        "citizenId"
      ]
    }
  ]
}
```

**NOTE**

- Add `accepted_namespace_list`


## GetServicesByAsID

### Expected Output

```sh
{
  "services": [
    {
      "active": true,
      "min_aal": 1.1,
      "min_ial": 1.1,
      "service_id": "AFLHeKQVLNQOkIOxoNid",
      "accepted_namespace_list": [
        "citizenId"
      ],
      "suspended": false
    },
    {
      "active": true,
      "min_aal": 2.2,
      "min_ial": 2.2,
      "service_id": "qvyfrfJRsfaesnDsYHbH",
      "accepted_namespace_list": [
        "citizenId"
      ],
      "suspended": false
    },
    {
      "active": true,
      "min_aal": 3.3,
      "min_ial": 3.3,
      "service_id": "JTFHqDoJRccWcikcJqnL",
      "accepted_namespace_list": [
        "citizenId"
      ],
      "suspended": false
    }
  ]
}
```

**NOTE**

- Add `accepted_namespace_list`


## CreateIdpResponse

### Parameter

```sh
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

```sh
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

```sh
{
  "reference_group_code": "aaaaa-bbbbb-ccccc-ddddd",
  "identity_namespace": "citizenId",
  "identity_identifier_hash": "c765a80f1ee71299c361c1b4cb4d9c36b44061a526348a71287ea0a97cea80f6",
  "node_id": "CuQfyyhjGcCAzKREzHmL"
}
```

**NOTE**

- Remove `hash_id`
- Add `reference_group_code`, `identity_namespace` and `identity_identifier_hash`
- Input `reference_group_code` or `identity_namespace`+`identity_identifier_hash` (able to input one or the other, if both, error)


## UpdateIdentity

### Parameter

```sh
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


## Remove These

- ClearRegisterIdentityTimeout 
- SetTimeOutBlockRegisterIdentity
- RegisterAccessor
- DeclareIdentityProof
- GetIdentityProof
- `identity_proof` in `response_list` when calling GetRequestDetail
