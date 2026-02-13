# HTTP API

Full REST API reference. All endpoints accept/return JSON unless noted.

## Public Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/config` | Server configuration (feature flags, limits) |
| `GET` | `/version` | Build info (version, git commit) |
| `GET` | `/qrcode?url=...&size=...` | Generate QR code PNG |
| `GET` | `/health` | Health check |

## Upload & File Endpoints

Authentication: session cookie or `X-PlikToken` header.

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/` | Quick upload: create upload + add file |
| `POST` | `/upload` | Create upload with options |
| `GET` | `/upload/{uploadID}` | Get upload metadata |
| `DELETE` | `/upload/{uploadID}` | Delete upload |
| `POST` | `/file/{uploadID}` | Add file (multipart) |
| `POST` | `/file/{uploadID}/{fileID}/{filename}` | Add file with known ID (stream mode) |
| `DELETE` | `/file/{uploadID}/{fileID}/{filename}` | Remove file |
| `GET` | `/file/{uploadID}/{fileID}/{filename}` | Download file |
| `HEAD` | `/file/{uploadID}/{fileID}/{filename}` | File metadata |
| `POST` | `/stream/{uploadID}/{fileID}/{filename}` | Stream upload |
| `GET` | `/stream/{uploadID}/{fileID}/{filename}` | Stream download |
| `GET` | `/archive/{uploadID}/{filename}` | Download all files as zip |

### Create Upload (POST /upload)

```json
{
    "ttl": 86400,
    "oneShot": false,
    "removable": true,
    "stream": false,
    "login": "foo",
    "password": "bar",
    "comments": "optional markdown"
}
```

Response:

```json
{
    "id": "TczL35OTIb3InNr6",
    "uploadToken": "50lGHbLEIrpJOl4uECddTI7pga...",
    "files": []
}
```

### Add File (POST /file/{uploadID})

Send as `multipart/form-data` with `file` field. The `X-UploadToken` header is required (returned from upload creation).

### Download File

The upload token is not required for public uploads. For password-protected uploads, provide HTTP Basic auth with the upload's login/password.

## Authentication Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/auth/google/login` | Get Google consent URL |
| `GET` | `/auth/google/callback` | Google OAuth callback |
| `GET` | `/auth/ovh/login` | Get OVH consent URL |
| `GET` | `/auth/ovh/callback` | OVH OAuth callback |
| `GET` | `/auth/oidc/login` | Get OIDC consent URL |
| `GET` | `/auth/oidc/callback` | OIDC callback |
| `POST` | `/auth/local/login` | Login `{ "login": "...", "password": "..." }` |
| `GET` | `/auth/logout` | Logout |

## User Endpoints

Requires authenticated session cookie.

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/me` | Current user info |
| `DELETE` | `/me` | Delete own account |
| `GET` | `/me/token` | List tokens (paginated) |
| `POST` | `/me/token` | Create upload token `{ "comment": "..." }` |
| `DELETE` | `/me/token/{token}` | Revoke token |
| `GET` | `/me/uploads` | List uploads (paginated) |
| `DELETE` | `/me/uploads` | Remove all uploads |
| `GET` | `/me/stats` | User statistics |

## Admin Endpoints

Requires admin session cookie.

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/user` | Create user |
| `GET` | `/user/{userID}` | Get user info |
| `POST` | `/user/{userID}` | Update user |
| `DELETE` | `/user/{userID}` | Delete user |
| `GET` | `/stats` | Server statistics |
| `GET` | `/users` | List all users (paginated) |
| `GET` | `/uploads` | List all uploads (paginated) |

## Pagination

Paginated endpoints accept these query parameters:

| Parameter | Default | Description |
|-----------|---------|-------------|
| `offset` | `0` | Skip N results |
| `limit` | `50` | Max results per page |
| `order` | `desc` | Sort order (`asc`/`desc`) |
| `sort` | `createdAt` | Sort field |
