# Architecture — Helm Chart (`charts/plik/`)

> Kubernetes deployment chart for Plik. For system-wide overview, see the root [ARCHITECTURE.md](../../ARCHITECTURE.md).

---

## Structure

```
charts/plik/
├── Chart.yaml                  ← Chart metadata (version set at release time via __VERSION__)
├── values.yaml                 ← All user-configurable values (annotated for helm-docs)
├── README.md.gotmpl            ← helm-docs template for generating README.md
├── README.md                   ← Auto-generated values reference (do not edit manually)
├── CHANGELOG.md                ← Keep-a-Changelog (update [Unreleased] before each release)
├── ARCHITECTURE.md             ← this file
└── templates/
    ├── _helpers.tpl            ← Template helpers (plik.fullname, plik.secretName, etc.)
    ├── configmap.yaml          ← Renders plikd.cfg from non-sensitive plikd.* values
    ├── secret.yaml             ← Renders Kubernetes Secret from secrets.* values
    ├── deployment.yaml         ← Deployment or StatefulSet (controlled by .Values.kind)
    ├── service.yaml            ← ClusterIP service on port 8080
    ├── ingress.yaml            ← Optional Ingress resource
    ├── pvc.yaml                ← PVC for file/db data (when persistence/dbPersistence enabled)
    └── NOTES.txt               ← Post-install instructions
```

---

## Key Design Decisions

### Config vs. Secrets separation

| Category | Source | Rendered to | Mechanism |
|---|---|---|---|
| Non-sensitive config | `plikd.*` in `values.yaml` | ConfigMap (`plikd.cfg`) | TOML config file |
| Sensitive credentials | `secrets.*` in `values.yaml` | Kubernetes Secret | `envFrom.secretRef` → env var override |

The server loads the config file first, then applies env var overrides via `PLIKD_` prefix + screaming snake case (e.g., `GoogleAPISecret` → `PLIKD_GOOGLE_API_SECRET`). Map-type fields like `DataBackendConfig` accept JSON and **merge** into the config file map.

### BYO Secret (existingSecret)

Set `secrets.existingSecret: "my-secret-name"` to skip Secret creation and reference an externally managed secret (Vault, Sealed Secrets, ESO). The `plik.secretName` helper in `_helpers.tpl` resolves the correct name everywhere.

### Persistence

Two independent PVCs:
- `persistence` — uploaded file data at `/home/plik/server/files`
- `dbPersistence` — SQLite database at `/home/plik/server/db`

Both default to `emptyDir` when disabled. For `StatefulSet` mode, volumes use `volumeClaimTemplates`.

### Versioning

Chart `version` and `appVersion` in `Chart.yaml` use `__VERSION__` placeholders, replaced at release time by `releaser/helm_release.sh` to match the app release tag.
