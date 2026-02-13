# Prometheus Metrics

Plik exposes Prometheus metrics via a dedicated `/metrics` endpoint.
The metrics are served using a dedicated Prometheus registry that also includes Go runtime and process collectors.

## Configuration

The metrics endpoint is served on a separate HTTP server. Modules can register additional metrics via the `Register()` method on `PlikMetrics`.

## Available Metrics

### HTTP

| Metric | Type | Labels | Description |
|---|---|---|---|
| `plik_http_request_total` | Counter | method, path, code | Count of HTTP requests |

### Uploads & Files

| Metric | Type | Description |
|---|---|---|
| `plik_uploads_count` | Gauge | Total number of uploads in the database |
| `plik_anonymous_uploads_count` | Gauge | Total number of anonymous uploads in the database |
| `plik_users_count` | Gauge | Total number of users in the database |
| `plik_files_count` | Gauge | Total number of files in the database |
| `plik_size_bytes` | Gauge | Total upload size in the database |
| `plik_anonymous_size_bytes` | Gauge | Total anonymous upload size in the database |

### Server Stats

| Metric | Type | Description |
|---|---|---|
| `plik_server_stats_refresh_duration_second` | Histogram | Duration of server stats refresh requests |
| `plik_last_stats_refresh_timestamp` | Gauge | Timestamp of the last server stats refresh |

### Cleaning

| Metric | Type | Description |
|---|---|---|
| `plik_cleaning_duration_second` | Histogram | Duration of cleaning runs |
| `plik_cleaning_removed_uploads` | Counter | Uploads removed by cleaning routine |
| `plik_cleaning_deleted_files` | Counter | Files deleted by cleaning routine |
| `plik_cleaning_deleted_uploads` | Counter | Uploads deleted by cleaning routine |
| `plik_cleaning_removed_orphan_files` | Counter | Orphan files removed by cleaning routine |
| `plik_cleaning_removed_orphan_tokens` | Counter | Orphan tokens removed by cleaning routine |
| `plik_last_cleaning_timestamp` | Gauge | Timestamp of the last server cleaning |

### Runtime

Go runtime and process metrics are also exposed via the standard `go_*` and `process_*` collectors.
