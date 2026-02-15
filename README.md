[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)
[![Build](https://github.com/root-gg/plik/actions/workflows/master.yaml/badge.svg)](https://github.com/root-gg/plik/actions/workflows/master.yaml)
[![Go Report](https://img.shields.io/badge/Go_report-A+-brightgreen.svg)](http://goreportcard.com/report/root-gg/plik)
[![Docker Pulls](https://img.shields.io/docker/pulls/rootgg/plik.svg)](https://hub.docker.com/r/rootgg/plik)
[![GoDoc](https://godoc.org/github.com/root-gg/plik?status.svg)](https://godoc.org/github.com/root-gg/plik)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](http://opensource.org/licenses/MIT)

Want to chat with us ? Telegram channel : https://t.me/plik_rootgg

# Plik

Plik is a scalable & friendly temporary file upload system — like WeTransfer, self-hosted.

### Features

- ☁️ Multiple storage backends (local, S3, OpenStack Swift, Google Cloud Storage)
- 🗄️ Multiple metadata backends (SQLite, PostgreSQL, MySQL)
- 🔑 Multiple authentication providers (Local, Google, OVH, OIDC)
- ⏱️ Configurable TTL with auto-cleanup
- 🔐 Password-protected uploads
- 💣 OneShot downloads (file deleted after first download)
- ⚡ Stream mode (uploader → downloader, nothing stored)
- 📦 Archive & encrypt from CLI (tar/zip, openssl/pgp)
- 📊 Prometheus metrics
- 🖥️ Modern Vue 3 web interface

### Third-party clients

   - [ShareX](https://getsharex.com/) Uploader : Directly integrated into ShareX
   - [plikSharp](https://github.com/iss0/plikSharp) : A .NET API client for Plik
   - [Filelink for Plik](https://gitlab.com/joendres/filelink-plik) : Thunderbird Addon to upload attachments to Plik

### Quick Start

```bash
# Docker
docker run -p 8080:8080 rootgg/plik

# From release
wget https://github.com/root-gg/plik/releases/download/1.3.8/plik-server-1.3.8-linux-amd64.tar.gz
tar xzvf plik-server-1.3.8-linux-amd64.tar.gz
cd plik-server-1.3.8-linux-amd64/server && ./plikd

# From source
git clone https://github.com/root-gg/plik.git
cd plik && make
cd server && ./plikd
```

Open web interface at [http://127.0.0.1:8080](http://127.0.0.1:8080)

### CLI Upload

```bash
plik myfile.txt
# or
curl --form 'file=@/path/to/file' http://127.0.0.1:8080
```

### Documentation

📖 **[Full Documentation](https://root-gg.github.io/plik/)** — guides, configuration reference, API docs, and more.

### How to Contribute

Contributions are welcome! See the [contributing guide](https://root-gg.github.io/plik/contributing) for development setup and build instructions.

### License

[MIT](LICENSE)
