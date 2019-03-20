# ABCI v2.0

## RegisterIdentity

### Parameter

```javascript
{
  "users": [
    {
      "reference_code": "aaaaa-bbbbb-ccccc-ddddd",
      "sid_hash": "c765a80f1ee71299c361c1b4cb4d9c36b44061a526348a71287ea0a97cea80f6",
      "ial": 3,
      "mode": 2, // allowed 1 (?), 2, 3
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
- Add `reference_code` property (string)
- Change `hash_id` to `sid_hash`
- Add `mode`
- Add `accessor_id`
- Add `accessor_public_key`
- Add `accessor_type`
- Add `request_id`
- Check for existing `sid`. If already associated to an `reference_code`, error.

## AddAccessorMethod

### Parameter

```javascript
{
  "reference_code": "aaaaa-bbbbb-ccccc-ddddd",
  "sid_hash": "c765a80f1ee71299c361c1b4cb4d9c36b44061a526348a71287ea0a97cea80f6",
  "accessor_id": "07938aa2-2aaf-4bb5-9ccd-33700581e870",
  "accessor_public_key": "-----BEGIN PUBLIC KEY-----\\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAhdKdvawPO8XXroiAGkxF\\nfLRCqvk4X2iAMStq1ADjmPPWhKgF/ssU9LBdHKHPPX1+NMOX29gOL3ZCxfZamKO6\\nAbODt1e0bVfblWWMq5uMwzNrFo4nKas74SLJwiMg0vtn1NnHU4QTTrMYmGqRf2WZ\\nIN9Iro4LytUTLEBCpimWM2hodO8I60bANAO0gI96BzAWMleoioOzWlq6JKkiDsj7\\n8EjCI/bY1T/v4F7rg2FxrIH/BH4TUDy88pIvAYy4nNEyGyr8KzMm1cKxOgnJI8On\\nwT8HrAJQ58T3HCCiCrKAohkYBWITPk3cmqGfOKrqZ2DI+a6URofMVvQFlwfYvqU6\\n5QIDAQAB\\n-----END PUBLIC KEY-----",
  "accessor_type": "accessor_type_2",
  "request_id": "edaec8df-7865-4473-8707-054dd0cffe2d"
}
```

**NOTE**

- Remove `accessor_group_id`
- Add `sid_hash` and `reference_code`
- Input `reference_code` or `sid_hash` (able to input one or the other, if both then error when sid is not in that reference_code?)

## UpgradeIdentityMode (New) (ชื่อนี้ดีไหม?)

```javascript
{
    "sid_hash": "c765a80f1ee71299c361c1b4cb4d9c36b44061a526348a71287ea0a97cea80f6",
    "mode": 3,
    "request_id": "edaec8df-7865-4473-8707-054dd0cffe2d"
}
```

## GetUID

```javascript
{
    "sid_hash": "c765a80f1ee71299c361c1b4cb4d9c36b44061a526348a71287ea0a97cea80f6"
}
```

### Output

```javascript
{
    "reference_code": "aaaaa-bbbbb-ccccc-ddddd"
}
```

## GetIdpNodes

```javascript
{
  "reference_code": "aaaaa-bbbbb-ccccc-ddddd",
  "sid_hash": "c765a80f1ee71299c361c1b4cb4d9c36b44061a526348a71287ea0a97cea80f6",
  "min_aal": 3,
  "min_ial": 3,
  "node_id_list": null
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

- Input `reference_code` or `sid_hash` (able to input one or the other, if both then error when sid is not in that reference_code?)
- Change `hash_id` to `sid_hash`
- Add `mode` to result of `GetIdpNodesInfo`

## RevokeAccessorMethod (ตอนนี้มันเป็นยังไง?)

### Parameter

```javascript
{
  "accessor_id_list": [
    "11d10976-aede-4ba0-9f44-fc0c96db1f32"
  ],
  "request_id": "e7dcf1c2-eea7-4dc8-af75-724cf86454ef"
}
```


## RevokeIdentity (New) (ชื่อนี้ดีไหม?)

### Parameter

```javascript
{
  "reference_code": "aaaaa-bbbbb-ccccc-ddddd",
  "sid_hash": "c765a80f1ee71299c361c1b4cb4d9c36b44061a526348a71287ea0a97cea80f6",
  "request_id": "edaec8df-7865-4473-8707-054dd0cffe2d"
}
```


## Remove These

- ClearRegisterIdentityTimeout
- SetTimeOutBlockRegisterIdentity
- RegisterAccessor
- DeclareIdentityProof
- GetIdentityProof
- `identity_proof` in `response_list` when calling GetRequestDetail