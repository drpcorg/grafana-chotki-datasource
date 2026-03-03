# Chotki Grafana Data Source (MVP)

Read-only Grafana data source plugin for `AggregatorService` (Chotki/Aggregator gRPC).

## Scope

- Only read RPC methods (strict allowlist):
  - `GetOwner`, `GetFullOwner`, `GetOwnerHits`, `GetOwnerMetadata`
  - `GetKey`, `GetKeyHits`, `ListKeys`
  - `GetOwnersWithBalance`, `GetAllOwnerIds`
  - `GetNodeCoreKey`, `ListNodeCoreKeys`
- Query model: `method + params` (no SQL/DSL in MVP)
- Query editor modes:
  - Builder (form fields by method)
  - Raw JSON
- Output format:
  - `table`
  - `stat`

## Query model

```json
{
  "mode": "rpc",
  "method": "ListKeys",
  "params": {
    "ownerId": "7e03c6f9-ede5-4225-9454-5bb34db55ce1",
    "limit": 100
  },
  "options": {
    "format": "table",
    "decodeIds": true,
    "decodeEnums": true,
    "decodeTimestamps": true,
    "limit": 200
  }
}
```

## Data source configuration

`jsonData`:

- `grpcAddress` (`host:port`)
- `insecure` (default: `true`)
- `timeoutMs` (default: `4000`)
- `defaultLimit` (default: `200`)
- `hardLimit` (default: `1000`)
- `decodeIds`, `decodeEnums`, `decodeTimestamps` (default: `true`)

`secureJsonData`:

- `authToken`
- `tlsCACert`
- `tlsClientCert`
- `tlsClientKey`

## Readability conversions

- `bytes` IDs (`ownerId`, `keyId`) -> UUID string (or base64 if `decodeIds=false`)
- Timestamps -> `time.Time` + RFC3339 string + unix seconds
- Enum/int labels:
  - `tier` -> `free/paid`
  - `mev_mode` -> `unset/enabled/disabled`
- `PublicKey.bytes` -> base64
- `ClientSpec` -> parse JSON if valid, else keep raw
- Arrays -> dual representation (`*_json` + `*_csv`)

Sensitive fields are not masked in MVP.

## Local development

### 1. Install dependencies

```bash
npm install
```

### 2. Build frontend in watch mode

```bash
npm run dev
```

### 3. Build backend binaries

```bash
mage -v
```

### 4. Start Grafana via docker-compose

```bash
npm run server
```

## Provisioning example

See:

- `provisioning/datasources/datasources.yml`

## Deployment

### Internal dev (unsigned)

- Enable unsigned plugins in Grafana:
  - `allow_loading_unsigned_plugins=drpc-chotki-datasource`
- Deploy artifact and restart Grafana.

### Internal production (signed private)

- Sign plugin:

```bash
npm run sign
```

- Publish versioned artifact to internal storage/registry.
- Canary rollout: one Grafana instance -> full pool.
- Rollback: previous artifact version + Grafana restart.

## Testing

### Backend unit tests

```bash
go test ./pkg/plugin/... -count=1
```

### Frontend checks

```bash
npm run typecheck
npm run test:ci
```

### E2E (optional)

```bash
npm run server
npm run e2e
```

## Notes

- `plugin.json` changes require Grafana restart.
- Plugin ID is fixed: `drpc-chotki-datasource`.
