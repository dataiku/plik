# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- `secret.yaml` now reads credentials exclusively from `secrets.*` values
- `deployment.yaml` uses `plik.secretName` helper for consistent Secret name resolution
  and adds `optional: true` on `envFrom.secretRef` so pods start cleanly without a Secret
- Sensitive fields removed from `plikd:` in `values.yaml` so they no longer leak into
  the ConfigMap-rendered `plikd.cfg`

### Added
- New top-level `secrets:` block in `values.yaml` to hold all sensitive credentials
  (`googleApiSecret`, `ovhApiKey`, `ovhApiSecret`, `oidcClientSecret`, `dataBackend`, `metadataBackend`)
- `secrets.existingSecret` — bring-your-own Secret support (replaces top-level `existingSecret`)
- `plik.secretName` Helm helper for consistent Secret name resolution
- `dbPersistence` — dedicated PVC for the SQLite metadata database
- Ingress template, post-install notes, Kubernetes deployment guide
- Explicit key ordering in `configmap.yaml` (fixes non-deterministic rendering)

### Removed
- Top-level `existingSecret` value (replaced by `secrets.existingSecret`)
- Sensitive fields from under `plikd.*` in `values.yaml`:
  `GoogleApiSecret`, `OvhApiKey`, `OvhApiSecret`, `OIDCClientSecret`,
  `DataBackendConfig.SecretAccessKey`, `DataBackendConfig.ApiKey`,
  `MetadataBackendConfig.Password`

> [!IMPORTANT]
> **Migration from previous chart versions**: All sensitive credentials must now be
> placed under the `secrets:` top-level block in `values.yaml`, not under `plikd:`.
