# Kubernetes (Helm)

> [!IMPORTANT]
> **Data Safety Warning**: For production deployments, we strongly advise using a **dedicated SQL database** (PostgreSQL or MySQL) for metadata, and enabling **persistence** (PVC) or using a **remote backend** (S3, GCS, Swift) for file storage to ensure no data is lost.

Plik provides a Helm chart to deploy it on Kubernetes.

The chart source is located in `charts/plik/`.

## Installation

### From Repository

```bash
helm repo add plik https://root-gg.github.io/plik
helm repo update
helm install plik plik/plik
```

### From Source
```bash
git clone https://github.com/root-gg/plik.git
cd plik/charts/plik
helm install plik .
```

## Configuration

The chart supports several modes of deployment. All Plik [configuration](./configuration.md) options are available under the `plikd` key in `values.yaml`.

### Persistence

By default, the chart uses `emptyDir` for storage. For production, you should enable persistence and choose the deployment kind.

#### StatefulSet (Recommended when using File Data Backend)

If you use the `file` backend, a `StatefulSet` with a `PersistentVolumeClaim` is recommended to ensure files are consistently stored.

```yaml
kind: StatefulSet
persistence:
  enabled: true
  size: 10Gi
```

#### Deployment (Recommended when using Cloud Data Backends)

If you use S3, GCS, or Swift, a standard `Deployment` is more appropriate.

```yaml
kind: Deployment
plikd:
  DataBackend: s3
  DataBackendConfig:
    Endpoint: "s3.amazonaws.com"
    Bucket: "my-bucket"
    # Credentials should be provided via secret
```

### Secrets Management

Sensitive values (API secrets, database passwords) are managed via a Kubernetes Secret. You can either provide them in `values.yaml` (which will be automatically extracted into a Secret and injected as environment variables) or specify an `existingSecret`.

```yaml
# Use an existing secret containing PLIKD_GOOGLE_API_SECRET, etc.
existingSecret: "my-custom-plik-secret"
```

The following environment variables are supported for secret injection:

| Environment Variable | Values Key |
|---------------------|------------|
| `PLIKD_GOOGLE_API_SECRET` | `plikd.GoogleApiSecret` |
| `PLIKD_OVH_API_SECRET` | `plikd.OvhApiSecret` |
| `PLIKD_OVH_API_KEY` | `plikd.OvhApiKey` |
| `PLIKD_OIDC_CLIENT_SECRET` | `plikd.OIDCClientSecret` |
| `PLIKD_DATA_BACKEND_CONFIG` | `plikd.DataBackendConfig` (sensitive keys as JSON) |
| `PLIKD_METADATA_BACKEND_CONFIG` | `plikd.MetadataBackendConfig` (sensitive keys as JSON) |

## Service and Ingress

The service is exposed via a `ClusterIP` by default. You can configure an `Ingress` in `values.yaml`.

```yaml
ingress:
  enabled: true
  className: nginx
  hosts:
    - host: plik.example.com
      paths:
        - path: /
          pathType: ImplementationSpecific
  tls:
    - secretName: plik-tls
      hosts:
        - plik.example.com
```

## Upgrading

When upgrading the chart, use:

```bash
helm repo update
helm upgrade plik plik/plik
```
