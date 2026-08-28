# Changelog

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
