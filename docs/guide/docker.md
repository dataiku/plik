# Docker Deployment

Plik provides multiarch Docker images for production deployments.

## Quick Start

```bash
docker run -p 8080:8080 rootgg/plik
```

Available tags:
- `latest` — Latest stable release
- `preview` — Latest release (including pre-releases like `-RC`)
- `__VERSION__` — Specific version (e.g. `1.4.0`, `1.4-RC1`)
- `dev` — Latest build from master branch

## Custom Configuration

Mount a custom `plikd.cfg` file:

```bash
docker run -p 8080:8080 \
  -v /path/to/plikd.cfg:/home/plik/server/plikd.cfg \
  rootgg/plik
```

See the [Configuration Guide](./configuration.md) for all available options.

## Persistent Storage

Mount a data directory for file storage:

```bash
docker run -p 8080:8080 \
  -v /data:/home/plik/server/files \
  rootgg/plik
```

## Docker Compose

Create a `docker-compose.yml`:

```yaml
version: "2"
services:
  plik:
    image: rootgg/plik:latest
    container_name: plik
    volumes:
      - ./plikd.cfg:/home/plik/server/plikd.cfg
      - ./data:/data
    ports:
      - 8080:8080
    restart: "unless-stopped"
```

Configure `plikd.cfg` to use the mounted volume:

```toml
DataBackend = "file"
[DataBackendConfig]
    Directory = "/data/files"

[MetadataBackendConfig]
    Driver = "sqlite3"
    ConnectionString = "/data/plik.db"
```

Start the container:

```bash
docker-compose up -d
```

## Environment Variables

You can override config values via environment variables:

```bash
docker run -p 8080:8080 \
  -e PLIKD_MAX_FILE_SIZE_STR=10GB \
  -e PLIKD_DEFAULT_TTL_STR=86400 \
  rootgg/plik
```

## Health Check

The Docker image includes a health check endpoint at `/health`:

```bash
curl http://localhost:8080/health
```

## Building Custom Images

To build from source:

```bash
make docker
```

This creates a `rootgg/plik:dev` image with the current codebase.

## Redeploying with a Custom Image

You can redeploy your instance with a custom image built from a Pull Request.

### Automated Deployment (GitHub Actions)

If you have configured the necessary secrets in your repository, you can trigger an automated deployment by commenting on a PR:

1. Comment `docker build` to build and push the PR image (`rootgg/plik:pr-{PR_NUMBER}`).
2. Comment `docker deploy` to deploy this image to your production server.

**Required GitHub Secrets:**
- `DEPLOY_HOST`: Production server IP/hostname.
- `DEPLOY_USER`: SSH user.
- `DEPLOY_SSH_KEY`: SSH private key.
- `DEPLOY_PATH`: Absolute path to the directory containing `docker-compose.yml` on the server.

### Manual Deployment

To manually redeploy with a specific image:

1. SSH into your server.
2. Update the image tag in your `docker-compose.yml`:
   ```yaml
   services:
     plik:
       image: rootgg/plik:pr-123  # Target PR number
   ```
3. Pull and restart:
   ```bash
   docker-compose pull plik
   docker-compose up -d plik
   ```
