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

## Real Grafana + dproxy chotki E2E

This flow validates the datasource against a real `chotki` gRPC server from the `dproxy` project.

### 1. Start only chotki in dproxy

```bash
cd ../dproxy
docker compose up -d chotki
docker compose ps chotki
```

Optional gRPC check:

```bash
grpcurl -plaintext localhost:9393 list
```

Expected service: `aggregator_api.AggregatorService`.

### 2. Prepare plugin Grafana integration env

```bash
cd /path/to/grafana-datasource-chotki/drpc-chotki-datasource
cp .env.integration.example .env.integration
```

Default values:

- `DPROXY_NETWORK_NAME=dproxy_keymanager`
- `CHOTKI_GRPC_ADDR=chotki:9393`

### 3. Build plugin artifact

```bash
npm run build
mage -v
```

If `mage` is not installed:

```bash
go install github.com/magefile/mage@latest
```

### 4. Start Grafana connected to dproxy network

```bash
docker compose -f docker-compose.yaml -f docker-compose.integration.yaml --env-file .env.integration up -d
docker compose ps
```

Verify Grafana container is attached to the external dproxy network:

```bash
docker network inspect ${DPROXY_NETWORK_NAME:-dproxy_keymanager} | rg drpc-chotki-datasource
```

### 5. Manual smoke in Grafana UI

Open [http://localhost:3000](http://localhost:3000):

1. Data sources -> `chotki datasource` -> `Save & test`.
2. Explore -> run `GetAllOwnerIds` in `table` and `stat`.
3. If owner IDs exist:
   - `GetOwner`, `GetFullOwner`, `GetOwnerHits`, `GetOwnerMetadata`.
4. Keys flow:
   - `ListKeys` -> take `key_id` -> `GetKey`, `GetKeyHits`.
5. NodeCore flow:
   - `ListNodeCoreKeys` -> if found -> `GetNodeCoreKey`.
6. Run `GetOwnersWithBalance`.

### 6. Negative checks

1. `ownerId=not-a-uuid` -> expect validation error `parameter "ownerId" must be UUID or base64 bytes`.
2. Timeout (deterministic): `docker pause dproxy-chotki-1`, run query, then `docker unpause dproxy-chotki-1` -> expect `context deadline exceeded` (`status=504` in Query Inspector/API).
3. `limit=999999` in `ListKeys`/`ListNodeCoreKeys` -> returned rows do not exceed datasource `hardLimit`.
4. Stop chotki (`docker compose stop chotki` in dproxy) and rerun query -> expect transport error mapped to `BadGateway` (`status=502` in API response).

### 7. Acceptance checklist

| Check | Expected |
| --- | --- |
| Save & test | `connected to AggregatorService` |
| GetAllOwnerIds | Works in `table` and `stat` |
| Owner pipeline | Owner methods work for valid ownerId |
| Key pipeline | `ListKeys -> GetKey -> GetKeyHits` works |
| NodeCore pipeline | `ListNodeCoreKeys -> GetNodeCoreKey` works if data exists |
| Converter output | UUID/time/enum/array fields are readable |
| Error mapping | invalid arg / timeout / unavailable mapped correctly |

If `GetAllOwnerIds` returns empty result, infrastructure is considered healthy and owner/key/nodecore scenarios are `not applicable` until data appears.

## Notes

- `plugin.json` changes require Grafana restart.
- Plugin ID is fixed: `drpc-chotki-datasource`.
