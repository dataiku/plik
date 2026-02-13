# Getting Started

## From Release

Download the latest release and run:

```bash
wget https://github.com/root-gg/plik/releases/download/__VERSION__/plik-__VERSION__-linux-amd64.tar.gz
tar xzvf plik-__VERSION__-linux-amd64.tar.gz
cd plik-__VERSION__/server
./plikd
```

Plik is now running at [http://127.0.0.1:8080](http://127.0.0.1:8080).

Edit `plikd.cfg` to adjust the configuration (ports, TLS, TTL, backends, etc.).

## From Source

Requires Go and Node.js:

```bash
git clone https://github.com/root-gg/plik.git
cd plik && make
cd server && ./plikd
```

## Docker

Plik provides multiarch Docker images for linux amd64/i386/arm/arm64:

```bash
docker run -p 8080:8080 rootgg/plik
```

Available tags:
- `rootgg/plik:latest` — latest release
- `rootgg/plik:{version}` — specific release
- `rootgg/plik:dev` — latest master commit

See the [Docker Deployment Guide](./docker.md) for custom configuration, persistent storage, and Docker Compose examples.
