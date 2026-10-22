# Changelog

## 1.1.1 (2026-10-22)

- Dependency updates within Grafana 13.2 compatibility (#118):
  - `@grafana/*` 13.2.0 → 13.2.4, `grafana-plugin-sdk-go` v0.296.5 → v0.297.0, Go 1.27.3.
  - create-plugin scaffold 7.0.2 → 7.15.1 (`@grafana/eslint-config` 10, `diagnostics-webpack-plugin`).
  - React kept on 19.2.x and Node on 24 to match what Grafana ships.

## 1.1.0 (2026-08-28)

- Synced `pkg/api` with the upstream `aggregator_api.proto` (dproxy, 2026-08-26):
  - 10 new RPCs, including read methods `GetAuthSnapshot`, `GetAuthSnapshots`, `GetPackagePools`.
  - New `Owner` fields exposed in `GetOwner` (`discounts_json`, `billing_version`) and
    `GetFullOwner`/`GetOwnerMetadata` (`discounts_json`).
- New query methods:
  - `GetAuthSnapshot` — owner/key auth snapshot with counters and package pools.
  - `GetAuthSnapshots` — batched snapshots via `refs` (JSON array or `ownerId:keyId` pairs) or
    parallel `ownerIds`/`keyIds` lists; per-ref errors are returned as an `error` column.
  - `GetPackagePools` — prepaid package pools with `remaining` (credited − spent) per tag.
- `CheckHealth` now uses the cheap `BlockHeight` RPC instead of the full-scan `GetAllOwnerIds`.
- Proto sources vendored into `pkg/api/proto` for reproducible regeneration.
- Dependency updates: `grafana-plugin-sdk-go` v0.290.0 → v0.296.4, `@grafana/plugin-e2e` → ^3.11.1.

## 1.0.1

Stable editor selectors for e2e tests.

## 1.0.0

Initial release.
