# Architecture — GitHub Actions (`.github/`)

> CI/CD workflows and automation for Plik. For system-wide overview, see the root [ARCHITECTURE.md](../ARCHITECTURE.md).

---

## Structure

```
.github/
├── workflows/
│   ├── tests.yaml              ← CI: lint, test, docs build on push/PR
│   ├── pages.yml               ← Deploy docs to gh-pages
│   ├── release.yaml            ← Build release archives, Docker images, Helm chart on release
│   ├── master.yaml             ← Post-merge actions on master
│   ├── docker-build-pr.yaml    ← Build Docker image on PR (triggered by comment)
│   └── docker-deploy-pr.yaml   ← Deploy PR image to staging (triggered by comment)
└── ARCHITECTURE.md             ← this file
```

---

## Workflows

### `tests.yaml` — CI Tests

Runs on every push and pull request. Steps:
1. Go lint (`make lint`)
2. Go tests (`make test`)
3. Docs build (`make docs`) — verifies VitePress builds without errors

### `pages.yml` — GitHub Pages (Docs)

Runs on push to `master` when `docs/**` or the workflow itself changes. Deploys VitePress documentation to the `gh-pages` branch via `peaceiris/actions-gh-pages`.

> [!NOTE]
> The `keep_files: true` flag preserves the Helm `index.yaml` on `gh-pages` (updated by the release workflow).

### `release.yaml` — Tagged Release Pipeline

Triggered when a GitHub release is created (tag push). Runs the full release pipeline:
1. Builds multi-arch Docker images
2. Builds release archives and client binaries
3. Packages the Helm chart and updates `index.yaml` on `gh-pages` (via `releaser/helm_release.sh`)
4. Uploads all artifacts (including `plik-helm-{version}.tgz`) to the GitHub release

See [releaser/ARCHITECTURE.md](../releaser/ARCHITECTURE.md) for the build details.

### `master.yaml` — Docker Dev Build

Runs on every push to `master` (only in the `root-gg` org). Builds multi-arch Docker images and pushes `rootgg/plik:dev` to Docker Hub via `make release-and-push-to-docker-hub`. This ensures the `dev` tag always reflects the latest `master` state.

### `docker-build-pr.yaml` — PR Docker Build

Triggered by a `docker build` comment on a PR. Builds a Docker image tagged `rootgg/plik:pr-{number}` and pushes it to Docker Hub. Reports back with a comment.

### `docker-deploy-pr.yaml` — PR Docker Deploy

Triggered by a `docker deploy` comment on a PR. Deploys the PR-specific Docker image to the staging instance at `plik.root.gg`. Reports back with a deployment confirmation comment.

---

## Helm Chart Release Flow

```mermaid
graph LR
    Release["GitHub Release created"] --> ReleaseWF["release.yaml workflow"]
    ReleaseWF --> Build["Build archives + Docker"]
    ReleaseWF --> HelmScript["releaser/helm_release.sh"]
    HelmScript --> Package["Package plik-helm-VERSION.tgz"]
    Package --> Assets["GitHub Release assets"]
    HelmScript --> UpdateIndex["Update index.yaml"]
    UpdateIndex --> GHPages["gh-pages branch"]
    GHPages --> HelmRepo["helm repo add plik<br/>https://root-gg.github.io/plik"]
```

The Helm chart version is unified with the app version — `Chart.yaml` `version` and `appVersion` are dynamically set to the release tag by `helm_release.sh` at release time.

Users install the chart via:
```bash
helm repo add plik https://root-gg.github.io/plik
helm install plik plik/plik
```

The chart source lives in `charts/plik/`. See the chart's `values.yaml` for all configuration options.
