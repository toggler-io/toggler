# package httpapi

## GET /api/ws

This is a Websocket endpoint that able to execute operation it receives.
The operation correspond with the HTTP API endpoint operation names.
For more you can check the swagger documentation

### operation IsFeatureEnabled

Get rollout feature flag status for a pilot

#### Example

request
```json
{
  "operation": "IsFeatureEnabled",
  "data": {
    "feature": "my-feature",
    "id":"public-pilot-uniq-id"
  }
}
```

response
```json
{"enrollment": true}
```

### operation IsFeatureGloballyEnabled

Get rollout feature flag status for global

#### Example

request
```json
{
  "operation": "IsFeatureEnabled",
  "data": {"feature": "my-feature"}
}
```

response
```json
{"enrollment": true}
```

## GET /api/client/config.json

This endpoint able to answer multiple feature flag state for a specific pilot.

### Example

request
```json
{
  "id": "public-pilot-uniq-id",
  "features": [
    "flag-name-a",
    "flag-name-b",
    "flag-name-c"
  ]
}
```

response
```json
{
  "states": {
    "flag-name-a": true,
    "flag-name-b": false,
    "flag-name-c": true
  }
}
```

## GET /api/release/is-feature-enabled.json

Get rollout feature flag status for a pilot

### Example

request
```json
{
  "feature": "my-feature",
  "id":"public-pilot-uniq-id"
}
```

response
```json
{"enrollment": true}
```

## GET /api/release/is-feature-globally-enabled.json

Get rollout feature flag status for global

### Example

request
```json
{"feature": "my-feature"}
```

response
```json
{"enrollment": true}
```
